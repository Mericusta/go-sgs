# go-sgs

## version 0.0.1

### Design Ideas

#### Drafts

- TODO: 事件驱动，消息驱动

#### Resource Model

- use `middleware` as server option to control resources model

- resource model 1: `1 - 1 - 3`
    - generally no-need `dispatcher`
    - 1 client -> 1 socket -> 3 goroutine: read/write/logic

- resource model 2: `1 - 1 - 2 - 1/n`
    - need `dispatcher`
    - 1 client -> 1 socket -> 2 goroutine: read/write -> logic: 1/n goroutine

- resource model 3: `1 - 1 - 1/n - 1/m`
    - need multi `dispatcher`
    - 1 client -> 1 socket -> 1 goroutine: read -> logic: 1/n goroutine -> write: 1/m goroutine


- resource model 4: `1 - 1/l - 1/m - 1/n`
    - need multi `dispatcher`
    - 1 client -> 1 socket -> 1/l goroutine: read -> logic: 1/m goroutine -> write: 1/n goroutine
    - Note: expect golang feature: read channel without blocking, like try lock -> try read channel
    
#### Call chain level

- level 0: os tcp socket
- level 1: specific server program
- level 2: recv/send goroutine
- level 3: logic goroutine

#### Recv Goroutine

- recv goroutine 接收消息的 goroutine

#### Send Goroutine

- send goroutine 发送消息的 goroutine

#### Logic Goroutine

- logic goroutine 业务逻辑的 goroutine

### Concepts

#### Protocol

> github.com/Mericusta/go-sgs/protocol

- protocol 协议格式，数据压缩/解压/加密/解密成 []byte 的算法

- 支持协议格式：
    - json
    - protobuf
    - TODO: MessagePack
    - TODO: bson

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

#### Dispatcher

> github.com/Mericusta/go-sgs/dispatcher

- TODO: dispatcher 分发器
    - receive Msg from recv goroutine
    - dispatch Msg to Handler and make Context
    - dispatch Msg to send goroutine by Linker
    - maybe different goroutine/program

#### Handler

- handler 处理器

#### Framework

> github.com/Mericusta/go-sgs/framework

- framework 框架，对外服务器的基本框架，封装底层细节

### Process

- client - server transport process
    - os: tcp socket -> read goroutine: unpack []byte, unmarshal -> logic goroutine: handler
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