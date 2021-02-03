package discovery

import (
	"context"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/plugin"
	"github.com/pole-group/pole/pojo"
	"github.com/pole-group/pole/server/storage"
	"github.com/pole-group/pole/utils"
)

const PluginDiscoveryStorage string = "PluginDiscoveryStorage"

type DiscoveryStorage interface {
	// SaveService 保存一个 Service 对象
	SaveService(ctx context.Context, service *Service, call func(err error, service *Service))
	// BatchUpdateService 批量更新 Service 对象
	BatchUpdateService(ctx context.Context, services []*Service, call func(err error, services []*Service))
	// RemoveService 删除一个服务, 会级联删除该服务下的所有相关Cluster、Instance资源信息
	RemoveService(ctx context.Context, service *Service, call func(err error, service *Service))
	// ReadService 获取一个服务
	ReadService(ctx context.Context, query *Service) error
	// SaveCluster
	SaveCluster(ctx context.Context, cluster *Cluster, call func(err error, cluster *Cluster))
	// BatchUpdateCluster
	BatchUpdateCluster(ctx context.Context, clusters []*Cluster, call func(err error, clusters []*Cluster))
	// RemoveCluster
	RemoveCluster(ctx context.Context, cluster *Cluster, call func(err error, cluster *Cluster))
	// ReadCluster 获取一个Cluster
	ReadCluster(ctx context.Context, query *Cluster) error
	// SaveInstance
	SaveInstance(ctx context.Context, instance *Instance, call func(err error, instance *Instance))
	// BatchUpdateInstance
	BatchUpdateInstance(ctx context.Context, instances []*Instance, call func(err error, instances []*Instance))
	// RemoveInstance
	RemoveInstance(ctx context.Context, instance *Instance, call func(err error, instance *Instance))
	// ReadBatchInstance 批量读取多个服务实例
	ReadBatchInstance(ctx context.Context, query *Instance) error
}

func InitDiscoveryStorage() (bool, error) {
	return plugin.RegisterPlugin(common.NewCtxPole(), &kvDiscoveryStorage{})
}

type kvDiscoveryStorage struct {
	memoryKv storage.KVStorage
	diskKv   storage.KVStorage
}

func (kvD *kvDiscoveryStorage) Init(ctx context.Context) error {
	var err error
	kvD.memoryKv, err = storage.NewKVStorage(ctx, storage.MemoryKv)
	if err != nil {
		return err
	}
	kvD.diskKv, err = storage.NewKVStorage(ctx, storage.BadgerKv)
	if err != nil {
		return err
	}
	return nil
}

func (kvD *kvDiscoveryStorage) Run() {
	// do nothing
}

// ReadService 根据
func (kvD *kvDiscoveryStorage) ReadService(ctx context.Context, query *Service) error {
	values, err := kvD.diskKv.ReadBatch(ctx, [][]byte{[]byte(query.name), []byte(ParseServiceClusterList(query.name))})
	if err != nil {
		return err
	}
	originS := new(pojo.Service)
	if err := proto.Unmarshal(values[0], originS); err != nil {
		return err
	}
	query.originService = originS
	// 读取所有的Cluster名称数据
	clusterNames := strings.Split(string(values[1]), ",")
	clusters := make([][]byte, len(clusterNames))
	for i, clusterName := range clusterNames {
		clusters[i] = []byte(clusterName)
	}
	values, err = kvD.diskKv.ReadBatch(ctx, clusters)
	if err != nil {
		return err
	}
	for i, value := range values {
		originCluster := new(pojo.Cluster)
		if err := proto.Unmarshal(value, originCluster); err != nil {
			return err
		}
		cluster := &Cluster{
			lock:          sync.RWMutex{},
			name:          clusterNames[i],
			originCluster: nil,
			repository:    nil,
		}
		query.Clusters.Put(cluster.name, cluster)
	}
	return nil
}

func (kvD *kvDiscoveryStorage) SaveService(ctx context.Context, service *Service, call func(err error, service *Service)) {
	saveV := service.originService
	key := []byte(service.name)
	value, err := proto.Marshal(saveV)
	if err != nil {
		call(err, nil)
		return
	}

	keys := make([][]byte, 2, 2)
	values := make([][]byte, 2, 2)
	keys[0] = key
	values[0] = value

	keys[1] = []byte(ParseServiceClusterList(service.name))
	values[1] = []byte(utils.Join(service.Clusters.Keys(), ","))

	// 保存该Service以及该Service下的所有Cluster名称信息
	if err := kvD.diskKv.WriteBatch(ctx, keys, values); err != nil {
		call(err, nil)
		return
	}
	call(nil, service)
}

func (kvD *kvDiscoveryStorage) BatchUpdateService(ctx context.Context, services []*Service, call func(err error, services []*Service)) {
	keys := make([][]byte, len(services), len(services))
	values := make([][]byte, len(services), len(services))
	for i, service := range services {
		saveV := service.originService
		key := []byte(service.name)
		value, err := proto.Marshal(saveV)
		if err != nil {
			call(err, nil)
			return
		}
		keys[i] = key
		values[i] = value
	}
	if err := kvD.diskKv.WriteBatch(ctx, keys, values); err != nil {
		call(err, nil)
		return
	}
	call(nil, services)
}

func (kvD *kvDiscoveryStorage) RemoveService(ctx context.Context, service *Service, call func(err error, services *Service)) {
	key := []byte(service.name)
	if err := kvD.diskKv.Delete(ctx, key); err != nil {
		call(err, nil)
		return
	}
	call(nil, service)
}

func (kvD *kvDiscoveryStorage) SaveCluster(ctx context.Context, cluster *Cluster, call func(err error, cluster *Cluster)) {
	saveV := cluster.originCluster
	key := []byte(cluster.name)
	value, err := proto.Marshal(saveV)
	if err != nil {
		call(err, nil)
		return
	}
	if err := kvD.diskKv.Write(ctx, key, value); err != nil {
		call(err, nil)
		return
	}
	call(nil, cluster)
}

func (kvD *kvDiscoveryStorage) ReadCluster(ctx context.Context, query *Cluster) error {
	return nil
}

func (kvD *kvDiscoveryStorage) BatchUpdateCluster(ctx context.Context, clusters []*Cluster, call func(err error, clusters []*Cluster)) {
	keys := make([][]byte, len(clusters), len(clusters))
	values := make([][]byte, len(clusters), len(clusters))
	for i, cluster := range clusters {
		saveV := cluster.originCluster
		key := []byte(cluster.name)
		value, err := proto.Marshal(saveV)
		if err != nil {
			call(err, nil)
			return
		}
		keys[i] = key
		values[i] = value
	}
	if err := kvD.diskKv.WriteBatch(ctx, keys, values); err != nil {
		call(err, nil)
		return
	}
	call(nil, clusters)
}

func (kvD *kvDiscoveryStorage) RemoveCluster(ctx context.Context, cluster *Cluster, call func(err error, cluster *Cluster)) {
	key := []byte(cluster.name)
	if err := kvD.diskKv.Delete(ctx, key); err != nil {
		call(err, nil)
		return
	}
	call(nil, cluster)
}

func (kvD *kvDiscoveryStorage) SaveInstance(ctx context.Context, instance *Instance, call func(err error, instance *Instance)) {
	kvop := utils.IF(instance.IsTemporary(), kvD.memoryKv, kvD.diskKv).(storage.KVStorage)
	key := []byte(instance.GetKey())

	//TODO 这里需要先更新对应集群的Cluster列表信息，然后在插入Instance数据信息

	value, err := proto.Marshal(instance.originInstance)
	if err != nil {
		call(err, nil)
		return
	}
	if err := kvop.Write(ctx, key, value); err != nil {
		call(err, nil)
		return
	}
	call(nil, instance)
}

func (kvD *kvDiscoveryStorage) ReadBatchInstance(ctx context.Context, query *Instance) error {
	return nil
}

func (kvD *kvDiscoveryStorage) BatchUpdateInstance(ctx context.Context, instances []*Instance, call func(err error, instances []*Instance)) {
	kvop := utils.IF(instances[0].IsTemporary(), kvD.memoryKv, kvD.diskKv).(storage.KVStorage)
	keys := make([][]byte, len(instances), len(instances))
	values := make([][]byte, len(instances), len(instances))
	for i, instance := range instances {
		key := []byte(instance.GetKey())
		value, err := proto.Marshal(instance.originInstance)
		if err != nil {
			call(err, nil)
			return
		}
		keys[i] = key
		values[i] = value
	}
	if err := kvop.WriteBatch(ctx, keys, values); err != nil {
		call(err, nil)
		return
	}
	call(nil, instances)
}

func (kvD *kvDiscoveryStorage) RemoveInstance(ctx context.Context, instance *Instance, call func(err error, instance *Instance)) {
	kvop := utils.IF(instance.IsTemporary(), kvD.memoryKv, kvD.diskKv).(storage.KVStorage)
	key := []byte(instance.GetKey())
	if err := kvop.Delete(ctx, key); err != nil {
		call(err, nil)
		return
	}
	call(nil, instance)
}

func (kvD *kvDiscoveryStorage) Name() string {
	return PluginDiscoveryStorage
}

func (kvD *kvDiscoveryStorage) Destroy() {
	kvD.memoryKv.Destroy()
	kvD.diskKv.Destroy()
}
