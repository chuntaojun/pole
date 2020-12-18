# Pole

[![Go Report Card](https://goreportcard.com/badge/github.com/Conf-Group/conf)](https://goreportcard.com/report/github.com/Conf-Group/conf)

### description

The Golang version of [alibaba/nacos](https://github.com/alibaba/nacos), and carried out other detailed designs

借鉴了Alibaba/Nacos的大体设计思路，并重新规划功能的实现以及布局

### TODO

#### PoleServer 以及 PoleSDK-Go
- [x] 支持`RSocket`、`Http`、`gRPC`作为通信层，其中`RSocket`为将来`Conf-Group`出品的`SDK`的默认通行协议，`Http`则为为了快速对接多语言而提供的额外协议实现
- [ ] 存储插件化，支持不同的插件底座，默认实现基于`SQLite`、`MySQL`、`Pebble`的三种存储
- [ ] 服务发现的健康检查机制全部由客户端实现，服务端只记录实例信息以及租约的检查
- [x] 一致性协议层的抽象实现
    - [ ] Raft协议的适配
    - [ ] AP协议的实现
        - [ ] Gossip协议 or Nacos的Distro协议
- [ ] Console的实现
- [ ] 配置内容的传输加密
- [x] 更简单的权限控制机制, 将SDK的权限和Console的权限分离
  - [ ] 控制台白名单机制(使用正则匹配模式)
  - [ ] RBAC权限控制实现
  - [ ] 客户端使用令牌机制进行授权
- [ ] 更加友好的运维体验
- [ ] 云原生的运维、部署体验
- [ ] 多种集群节点发现的模式，除了`Alibaba/Nacos`的文件以及地址中心之外，额外基于`Kubernetes-Client-Api`的集群节点管理方式
- [ ] 多数据中心模式部署（有点难度，需要时间考虑设计的问题）
- [ ] 支持自身限制运行期间所需要的资源
- [ ] ...（待思考）

#### 周边生态组件
- [ ] ...（待处理）
