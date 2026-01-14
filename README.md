# go-sgs

## Tips

- actor 树结构，通过 ctx 管理生命周期，通过父子级控制 actor 的创建和销毁
- 实践中可能会到 CPU 打满的问题，所以需要想办法确定在出现这种情况时，哪个 actor 最活跃
- 基于 hybrid-rpc 的想法或者 MS 的 Orleans 框架的设计思想，指定 topic method args，向分布式服务器网络中任意逻辑节点发送消息
- 需要一个 topic 服务发现中心
- 需要一个可以从静态语法上得到目标 topic 的 method 的函数形式的实现，类似 grpc，示例：local.TopicManager.Topic.Method(args...)
