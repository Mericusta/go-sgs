# go-sgs

## version 0.0.1

### Definition

#### Global Definition 全局定义

##### Call Chain Level 调用链层级

- OS 系统层
- Framework 框架层
- Application 应用层

##### Goroutine Type 协程类型

- Send-Goroutine 发送协程
- Recv-Goroutine 接收协程
- Logic-Goroutine 逻辑协程

#### OS Level Definition 系统层定义

##### TCP Socket Packet tcp 套接字数据包

- tcp 套接字数据传递的最小单位
- tcp 套接字数据包支持自定义格式，默认格式为 TLV
- TLV 格式 tcp 套接字数据包
```
┌─────┬────────┬───────┐
│ Tag │ Length │ Value │
├─────┼────────┼───────┤
│  4  │   4    │       │
└─────┴────────┴───────┘
```

##### TCP Socket Message tcp 套接字消息

- `tcp socket packet` 的运行时定义
- 通常指**一个数据交换时的唯一标识和一个支持某种数据交换协议的结构体**
- 通常定义在应用层，并且给唯一标识绑定一个结构体的构造函数

#### Framework Level Definition 框架层定义

##### Protocol 数据交换协议

> github.com/Mericusta/go-sgs/protocol

- 做数据交换时，`运行时的内存数据`和 `[]byte` 相互转换的算法
    - TODO: 无法做 model 隔离，因为必须要导出
- `ProtocolID` 做数据交换时的标识定义
- `ProtocolMsg` 做数据交换时的结构定义

- 数据编码：将`运行时的内存数据`通过压缩/加密等手段生成 `[]byte` 数据
- 数据解码：将 `[]byte` 数据通过解压/解密等手段生成`运行时的内存数据`

- 支持数据交换格式：
    - json
    - protobuf
    - TODO: MessagePack
    - TODO: bson

##### Packer 打包器

> github.com/Mericusta/go-sgs/packer

- 持有一个 `net.Conn` 对象（视为 `OS` 层 tcp socket 的实例）
- 从 `OS` 中收发 `tcp socket packet`，根据 `OS` 进行差异化实现
- 打包：将 `运行时的内存数据` 按某种 `protocol` 编码成 `[]byte`，再按指定方式填充生成 `tcp socket packet`
- 解包：将 `tcp socket packet` 按指定方式读取成 `[]byte`，再按某种 `protocol` 解码成的 `运行时的内存数据`

##### Linker 链接器

> github.com/Mericusta/go-sgs/linker

- 持有一个 `packer` 对象
- 提供 `recv-goroutine` 接收协程的执行函数
- 提供 `send-goroutine` 发送协程的执行函数
- 持有 `recv-channel` 对象，接收其他协程的事件
- 持有 `send-channel` 对象，向其他协程发送事件

##### Event 事件

> github.com/Mericusta/go-sgs/event

- 包装 `tcp 套接字` 消息，用以在不同协程之间传递
- `logic-goroutine` 和其他协程交互的数据对象
- `logic-goroutine` 从 `recv-channel` 中接收来自 `recv-goroutine` 的数据
- `logic-goroutine` 从 `event-channel` 中接收来自其他协程的数据

##### Handler 处理器

- 通常指一个定义在应用层的符合类型的函数
- 通常需要一个标识来选择执行哪个处理器，标识可以为 `ProtocolID` 等
- 处理器按照上下文是否带有 linker 分为两种
    - 上下文不带有 linker：服务层处理器
    - 上下文带有 linker：应用层处理器

##### Dispatcher 分发器

> github.com/Mericusta/go-sgs/dispatcher

- 持有一个 `linker` 对象
- 通过 `recv-channel` 对象和 `recv-goroutine` 交互
- 通过 `send-channel` 对象和 `send-goroutine` 交互
- 持有 `event-channel` 对象和其他协程交互
- 提供 `logic-goroutine` 逻辑协程的执行函数

- 是否应该处理 `send-channel`/`recv-channel` 剩余的数据？
    - 不应该，理由如下：
    - 框架层到应用层的抽象结构和调用链是栈结构
    - 当远端应用层主动退出时，如主动离线：会从栈顶依次退出对应的抽象层（应用层，框架层，OS层），执行对应层次的逻辑，此时无法处理 send/recv 中的内容，因为执行的上下文不存在了
    - 当远端应用层被动退出时，如远端断网：会从栈底依次向栈顶退出，此时可以处理 send/recv 中的内容，因为执行的主体仍存在
    - 当本地应用层主动退出时，如踢掉客户端：会从栈顶依次退出对应的抽象层，执行对应层次的逻辑，此时无法处理 send/recv 中的内容，因为执行的主体不存在了
    - 当本地应用层被动退出时，如本地断网：会从栈底依次向栈顶退出，此时可以处理 send/recv 中的内容，因为执行的主体仍存在
    保险起见，都不处理

---

##### Framework 框架

> github.com/Mericusta/go-sgs/framework

- framework 框架，对外服务的基本框架，封装底层细节，控制运行时程序资源分配

##### Acceptor

> github.com/Mericusta/go-sgs/acceptor

- acceptor 接收器，产生 net.Conn 的方式
    - 服务器接收器：通过 net.Listener.Accept() 方法产生 net.Conn
    - 客户端接收器：通过 net.DialTimeout() 方法产生 net.Conn

- 一个 framework 可以同时支持多个接收器
- Acceptor 关闭不代表 net.Conn 需要关闭

##### Middleware

> middleware 指的是针对某一种行为的中间件 
> middleware 作为连接媒介，必须定义在某个 concept 中

###### Handle Middleware

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

---

### Design Ideas 设计理念

#### Drafts 草稿

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
- TODO: 运行时数据结构应当和交换协议的数据结构区分开，并建立自动化映射关系
- TODO: panic 池/中间件，提供应用层 panic recover 的能力
- TODO: pack/unpack 提供 []byte 缓存或使用 sync.Pool，减少重新分配内存的频率

#### Custom TCP Socket Packet 自定义套接字数据包

- 通过实现`打包器`的接口，来实现自定义套接字数据包
- 通过编译选项，来编译自定义的套接字数据包

#### Custom Protocol 自定义数据交换协议

- 通过实现函数 `func Marshal(any) ([]byte, error)` 和 `func Unmarshal(ProtocolID, []byte) (any, error)` 来实现自定义数据交换协议

#### Gourtine Resources 协程资源

- Recv-Goroutine 接收消息的 goroutine
    - 通过 `packer` 接收 `tcp socket` 收到的消息，并处理套接字异常
    - 将消息转换为 `event`，并通过 `recv-channel` 发送给其他协程
- Send-Goroutine 发送消息的 goroutine
    - 通过 `send-channel` 接收 `event`，并转换为消息
    - 通过 `packer` 向 `tcp socket` 发送消息，并处理套接字异常
- Logic Goroutine 处理逻辑的 goroutine
    - 执行 `handler`，为其提供上下文环境
    - 根据 `recv-channel` 或 `event-channel` 的状态来控制 `send-goroutine` 或 `recv-goroutine` 的结束
    - 不一定只由 `recv goroutine` 来触发，`logic goroutine` 本身是可以由**数据驱动**的（比如每隔一段时间主动推送消息或者接收到其他服务器推送给用户的消息）
        - 但**数据驱动**和 `dispatcher` 不好结合在一起，因为要**数据驱动**是独占 `logic goroutine` 的，而 `dispatcher` 的目的是共享 `logic goroutine`
        - 数据驱动的数据中带上 logic goroutine 的上下文就可以实现数据隔离，logic goroutine 共享
        - **数据驱动**独占 `logic goroutine` 可以转化为 `dispatcher` 独占 `logic goroutine` 并监听**数据驱动**

#### Resource Model 资源模型

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
    - Note: expect golang feature: recv-channel without blocking, like try lock -> try recv-channel

#### Call Chain Level 调用链层级

- 系统层 os
    - level 0: tcp socket
- 框架层 framework
    - level 1: connection send/recv goroutine
    - level 2: link send/recv goroutine
    - level 3: dispatcher logic goroutine
- 应用层 application
    - level 4: user logic goroutine
    - level 5: handler

- 框架层和应用层要有互相控住的方式，框架层 tcp socket 断开之后要控制应用层退出，应用层退出之后要断开框架层 tcp socket
- TODO: 框架层和应用层的退出应当有序

### Process

- client - server transport process
    - os: tcp socket -> recv goroutine: unpack []byte, unmarshal -> logic goroutine: handler
    ```
    ┌──────────────┬────────────────────────────────────┬─────────────────────┬─────────────────────────────┬──────────────────────────┐
    │      OS      │       recv-goroutine: packer       │   recv-goroutine    │ logic-goroutine: dispatcher │ logic-goroutine: handler │
    ├──────────────┼───────────────┬────────────────────┼─────────────────────┼─────────────────────────────┼──────────────────────────┤
    │  TCP Socket  │ unpack []byte │ unmarshal protocol │ recv-channel <- Msg │     Msg <- recv-channel     │        handle Msg        │
    └──────────────┴───────────────┴────────────────────┴─────────────────────┴─────────────────────────────┴──────────────────────────┘
    ```
    - logic goroutine: handler -> send goroutine: pack []byte, marshal -> os: tcp socket
    ```
    ┌──────────────────────────┬─────────────────────────────┬─────────────────────┬────────────────────────────────┬────────────┐
    │ logic-goroutine: handler │ logic-goroutine: dispatcher │   send-goroutine    │     send-goroutine: packer     │     OS     │
    ├──────────────────────────┼─────────────────────────────┼─────────────────────┼──────────────────┬─────────────┼────────────┤
    │         make Msg         │     send-channel <- Msg     │ Msg <- send-channel │ marshal protocol │ pack []byte │ TCP Socket │
    └──────────────────────────┴─────────────────────────────┴─────────────────────┴──────────────────┴─────────────┴────────────┘
    ```

- end process
    - from server:
        - server -> close listener -> close all link tcp socket connection -> cancel logic goroutine
    - from client:
        - TODO: client -> 

- link end process:
    - close link tcp socket connection
        - recv goroutine receive, then close recv-channel and end recv-goroutine
        - in server link, logic-goroutine will end by context canceler
        - in client link, logic-goroutine will
