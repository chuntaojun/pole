# Discovery 对象模型的设计

pole-discovery模块，对于服务发现实例的模型设计，是按照alibaba/nacos的设计来做的

```
service -> cluster -> instance
```

每个对象的`proto`

```protobuf
//服务
message Service {
  string serviceName = 1;
  string group = 2;
  ServiceMetadata metadata = 3;
  double protectThreshold = 5;
  string selector = 6;
}

//服务的元数据信息
message ServiceMetadata {
  map<string, string> metadata = 1;
}

enum CheckType {
  TCP = 0;
  HTTP = 1;
  CUSTOMER = 2;
  HeartBeat = 3;
}

//服务下的虚拟集群，可以利用Cluster做Region的事情，比如一个A服务部署在了SZ、GZ、SH、HZ，那么A服务的Cluster可以有Cluster_SZ、Cluster_SH、Cluster_GZ......
message Cluster {
  string serviceName = 1;
  string group = 2;
  string clusterName = 3;
  ClusterMetadata metadata = 5;
}
//服务下的虚拟集群的元数据信息
message ClusterMetadata {
  map<string, string> metadata = 1;
}

//实例，一个服务下有多个实例
message Instance {
  string serviceName = 1;
  string group = 2;
  string ip = 3;
  int64 port = 4;
  string clusterName = 5;
  double weight = 6;
  InstanceMetadata metadata = 7;
  bool ephemeral = 8;
  bool enabled = 9;
  CheckType healthCheckType = 10;
}

message InstanceMetadata {
  map<string, string> metadata = 1;
}
```

其中，每一个`cluster`下的`instance`，要么都为持久化实例，要么都为非持久化实例。