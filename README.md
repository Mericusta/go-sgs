# go-sgs

## version 0.0.1

### Design Ideas

#### Drafts

- TODO: 事件驱动，消息驱动
- TODO: 中间件：控制资源模型
- TODO: 中间件：隔离框架层和应用层上下文传递
- TODO: 中间件：支持断点调试（断点会阻塞进程）
- Link
    - 相同 linker，在经过不同编译条件的情况下，可以处理不同格式的 packet
    - [DEPRECATED]相同 linker，在不经过编译的情况下，可以处理不同格式的 packet
        - 在没有“额外信息（如何处理 packet）”的情况下不知道 packet 的处理格式
        - 提前发包告知
        - 每个包内告知
- TODO: 实际情况中，不同的服务器需要的指标不同，比如 gate 服务器需要保持大量 tcp 链接收发消息，game 服务器需要做大量逻辑处理等，同一套 Framework 是否真的能 hold 住所有的情况？考虑按性能指标做出特异化区分，比如 game 做逻辑处理就用空间换时间，gate 做 tcp 套接字处理就要做到容易被控制，出现异常容易察觉，恢复等

#### Resource Model

- use `middleware` as server option to control resources model

- resource model 1: `1 - 1 - 3`
    - generally no-need `dispatcher`, or `dispatcher` is logic goroutine
    - 1 client -> 1 socket -> 3 goroutine: recv/send/logic

- resource model 2: `1 - 1 - 2 - 1/n`
    - need logic `dispatcher`, n recv/send goroutine share one dispatcher
    - 1 client -> 1 socket -> 2 goroutine: recv/send -> logic: 1/n goroutine

- resource model 3: `1 - 1 - 1/n - 1/m`
    - need multi kinds `dispatcher`, logic dispatcher and send dispatcher
    - 1 client -> 1 socket -> 1 goroutine: recv -> logic: 1/n goroutine -> send: 1/m goroutine


- resource model 4: `1 - 1/l - 1/m - 1/n`
    - need multi kinds `dispatcher`, logic dispatcher, send dispatcher, recv dispatcher
    - 1 client -> 1 socket -> 1/l goroutine: recv -> logic: 1/m goroutine -> send: 1/n goroutine
    - Note: expect golang feature: recv channel without blocking, like try lock -> try recv channel

#### Call chain level

- level 0: tcp socket os
- level 1: connection send/recv goroutine
- level 2: link send/recv goroutine
- level 3: dispatcher logic goroutine
- level 4: user logic goroutine
- level 5: handler

level 4~5 是应用层
level 1~3 是框架层
level 0 是系统层

#### Recv Goroutine

- recv goroutine 接收消息的 goroutine

#### Send Goroutine

- send goroutine 发送消息的 goroutine

#### Logic Goroutine

> 不一定只由 `recv goroutine` 来触发，`logic goroutine` 本身是可以由**数据驱动**的（比如每隔一段时间主动推送消息或者接收到其他服务器推送给用户的消息）
> 但**数据驱动**和 `dispatcher` 不好结合在一起，因为要**数据驱动**是独占 `logic goroutine` 的，而 `dispatcher` 的目的是共享 `logic goroutine`
>   - 数据驱动的数据中带上 logic goroutine 的上下文就可以实现数据隔离，logic goroutine 共享
> **数据驱动**独占 `logic goroutine` 可以转化为 `dispatcher` 独占 `logic goroutine` 并监听**数据驱动**

- logic goroutine 业务逻辑的 goroutine
    - 被动接收，从 recv goroutine 来
    - 主动发送，往 send goroutine 去
- 考虑引入优先级 channel：
    - 优先级1：接收中断 > 接收消息 > 发送中断 > 主动发送
        - 被动的优先级高于主动
        - 针对主动的 logic goroutine，中断的优先级更高
        - 针对被动的 logic goroutine，中断的优先级更低

- 1 被动接收
- 2 被动结束
- 3 主动结束
- 4 主动发送

### Concepts

#### Framework Level Concepts

##### Framework

> github.com/Mericusta/go-sgs/framework

- framework 框架，对外服务的基本框架，封装底层细节

##### Dispatcher

> github.com/Mericusta/go-sgs/dispatcher

- TODO: dispatcher 分发器
    - 执行`逻辑协程`
    - 监听`发送通道`，主动发送消息，和 `send goroutine` 交互
    - 监听`接收通道`，被动接收消息，和 `recv goroutine` 交互

- 是否应该处理 send/recv 剩余的数据？
    - 不应该，理由如下：
    ```
    框架层到应用层的抽象结构和调用链是栈结构，从建立 tcp socket 开始抽象层次依次为：
        - level 0: tcp socket os
        - level 1: connection send/recv goroutine
        - level 2: link send/recv goroutine
        - level 3: dispatcher logic goroutine
        - level 4: user logic goroutine
        - level 5: handler 栈顶
    level 4~5 是应用层
    level 1~3 是框架层
    level 0 是系统层
    - 当远端应用层主动退出时，如主动离线：会从栈顶依次退出对应的抽象层，执行对应层次的逻辑，此时无法处理 send/recv 中的内容，因为执行的主体不存在了
    - 当远端应用层被动退出时，如远端断网：会从栈底依次向栈顶退出，此时可以处理 send/recv 中的内容，因为执行的主体仍存在
    - 当本地应用层主动退出时，如踢掉客户端：会从栈顶依次退出对应的抽象层，执行对应层次的逻辑，此时无法处理 send/recv 中的内容，因为执行的主体不存在了
    - 当本地应用层被动退出时，如本地断网：会从栈底依次向栈顶退出，此时可以处理 send/recv 中的内容，因为执行的主体仍存在
    保险起见，都不处理
    ```

##### Protocol

> github.com/Mericusta/go-sgs/protocol

- protocol 协议格式，数据压缩/解压/加密/解密成 []byte 的算法
    - TODO: 无法做 model 隔离，因为必须要导出

- 支持协议格式：
    - json
    - protobuf
    - TODO: MessagePack
    - TODO: bson

#### Acceptor

> github.com/Mericusta/go-sgs/acceptor

- acceptor 接收器，产生 net.Conn 的方式
    - 服务器接收器：通过 net.Listener.Accept() 方法产生 net.Conn
    - 客户端接收器：通过 net.DialTimeout() 方法产生 net.Conn

- 一个 framework 可以同时支持多个接收器
- Acceptor 关闭不代表 net.Conn 需要关闭

#### Connector

> github.com/Mericusta/go-sgs/connector

- connector 连接器
    - 连接框架层和系统层，负责 tcp socket packet 的收发
    - 负责将按照某种协议格式压缩/加密的 []byte 按指定方式打包成 tcp socket packet
    - 负责将 tcp socket packet 解包成 []byte 并按照某种协议格式解压/解密
    
#### Link

> github.com/Mericusta/go-sgs/link

- link 链接，tcp 链接的抽象
    - 提供 recv goroutine 的执行逻辑
    - 提供 send goroutine 的执行逻辑
    - 对接 logic goroutine 的执行逻辑

#### Event

> github.com/Mericusta/go-sgs/event

- `recv/send goroutine` 和 `logic goroutine` 传递消息的载体

#### Handler

> handle 行为：通过 protocol ID 查找函数回调（称为 handler）并执行的过程
> handle 行为的执行体

- handler 处理器，分为两类
    - 服务层处理器：不存在用户上下文
    - 用户层处理器：存在用户上下文，可能分为多种类型的用户（客户端用户，服务器用户等）

#### Middleware

> middleware 指的是针对某一种行为的中间件 
> middleware 作为连接媒介，必须定义在某个 concept 中

##### Handle Middleware

> github.com/Mericusta/go-sgs/dispatcher
> 针对 dispatcher 的 handle 行为的中间件

- handle 中间件
    - 流程控制
    - 多个中间件：
        - 层层传递？
            - Framework 层如何知道层级之间的关系？
        - 平级传递？
            - 中间件有相互依赖如何处理？（MiddlewareA -> MiddlewareB 而 MiddlewareB -x-> MiddlewareA）
        - 中间件排序？
            - 添加中间件的时候如何知道其他中间件的信息？
        - 通过 protocol ID 把消息路由到不同的 middleware 上去？
- TODO: 在中间件中，“在应用层容器中”查找唯一标识的数据，会遇到并发性能瓶颈

### Process

- client - server transport process
    - os: tcp socket -> recv goroutine: unpack []byte, unmarshal -> logic goroutine: handler
    ```
    ┌──────────────┬────────────────────────────────────┬─────────────────────┬─────────────────────────────┬──────────────────────────┐
    │      OS      │     recv goroutine: connector      │   recv goroutine    │ logic goroutine: dispatcher │ logic goroutine: handler │
    ├──────────────┼───────────────┬────────────────────┼─────────────────────┼─────────────────────────────┼──────────────────────────┤
    │  TCP Socket  │ unpack []byte │ unmarshal protocol │ recv channel <- Msg │     Msg <- recv channel     │        handle Msg        │
    └──────────────┴───────────────┴────────────────────┴─────────────────────┴─────────────────────────────┴──────────────────────────┘
    ```
    - logic goroutine: handler -> send goroutine: pack []byte, marshal -> os: tcp socket
    ```
    ┌──────────────────────────┬─────────────────────────────┬─────────────────────┬────────────────────────────────┬────────────┐
    │ logic goroutine: handler │ logic goroutine: dispatcher │   send goroutine    │   send goroutine: connector    │     OS     │
    ├──────────────────────────┼─────────────────────────────┼─────────────────────┼──────────────────┬─────────────┼────────────┤
    │         make Msg         │     send channel <- Msg     │ Msg <- send channel │ marshal protocol │ pack []byte │ TCP Socket │
    └──────────────────────────┴─────────────────────────────┴─────────────────────┴──────────────────┴─────────────┴────────────┘
    ```

- end process
    - from server:
        - server -> close listener -> close all link tcp socket connection -> cancel logic goroutine
    - from client:
        - TODO: client -> 

- link end process:
    - close link tcp socket connection
        - recv goroutine receive, then close recv channel and end recv goroutine
        - in server link, logic goroutine will end by context canceler
        - in client link, logic goroutine will