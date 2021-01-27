# Pole

[![Go Report Card](https://goreportcard.com/badge/github.com/pole-group/pole)](https://goreportcard.
com/report/github.com/pole-group/pole)

### description

The Golang version of [alibaba/nacos](https://github.com/alibaba/nacos), and carried out other detailed designs

借鉴了Alibaba/Nacos的大体设计思路，并重新规划功能的实现以及布局

### TODO

#### PoleServer 以及 PoleSDK-Go
- [x] 支持`RSocket`、`Http`、`gRPC`作为通信层，其中`RSocket`为将来`pole-group`出品的`SDK`的默认通行协议，`Http`则为为了快速对接多语言而提供的额外协议实现
- [x] 存储插件化，支持不同的插件底座，默认实现基于`Memory Map`、`Badger`的两种KV存储
    - [x] `key-value`形式的存储机制实现
    - [ ] 关系型数据库的存储机制实现
- [ ] 服务发现的健康检查机制全部由客户端实现，服务端只记录实例信息以及租约的检查
    - [x] Agent 模式的健康检查
        - [x] TCP 端口探测模式
        - [x] Http 接口访问模式
        - [x] 用户自定义检查函数模式
    - [ ] Server 进行租约检查
- [x] 一致性协议层的抽象实现
    - [ ] Raft协议的适配, [pole-group/lraft](https://github.com/pole-group/lraft) 项目，模范 sofa-jraft
    - [ ] AP协议的实现, Gossip协议 or Nacos的Distro协议
- [ ] Console的实现
- [ ] 配置内容的传输加密
- [x] 更简单的权限控制机制, 将SDK的权限和Console的权限分离
  - [ ] 权限系统抽象
  - [ ] 控制台白名单机制(使用正则匹配模式)
  - [ ] RBAC权限控制实现
  - [ ] 客户端使用令牌机制进行授权
- [ ] 更加友好的运维体验
- [ ] 云原生的运维、部署体验
- [ ] 多种集群节点发现的模式
    - [x] `Alibaba/Nacos`文件以及地址中心，
    - [x] 基于`Kubernetes-Client-Api`的集群节点管理方式
- [ ] 多数据中心模式部署（有点难度，需要时间考虑设计的问题）
- [ ] 支持自身限制运行期间所需要的资源
- [ ] ...（待思考）

#### 周边生态组件
- [ ] ...（待处理）
