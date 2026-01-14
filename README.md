# go-sgs

## Tips

- 在开始之前，先思考一下，是什么理由，促使你一定要用到多协程处理？
  - 单协程绑定单线程 -> 纯单线程逻辑
  - 少量协程绑定单线程 -> 无内核级线程切换
  - 大量协程使用多线程 -> 传统 GMP 模型
  - 多进程，多线程，多协程，分布式，HPA
  - ...
- actor 树结构，通过 ctx 管理生命周期，通过父子级控制 actor 的创建和销毁
- 实践中可能会到 CPU 打满的问题，所以需要想办法确定在出现这种情况时，哪个 actor 最活跃
- 基于 hybrid-rpc 的想法或者 MS 的 Orleans 框架的设计思想，指定 topic method args，向分布式服务器网络中任意逻辑节点发送消息
- 需要一个 topic 服务发现中心
- 需要一个可以从静态语法上得到目标 topic 的 method 的函数形式的实现，类似 grpc，示例：local.TopicManager.Topic.Method(args...)
