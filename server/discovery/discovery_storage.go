package discovery

import (
	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/plugin"
	"github.com/pole-group/pole/server/storage"
)

const PluginDiscoveryStorage string = "PluginDiscoveryStorage"

type DiscoveryStorage interface {
	SaveService(service *Service, call func(err error, service *Service))

	BatchUpdateService(services []*Service, call func(err error, services []*Service))

	RemoveService(service *Service, call func(err error, service *Service))

	SaveCluster(cluster *Cluster, call func(err error, cluster *Cluster))

	BatchUpdateCluster(clusters []*Cluster, call func(err error, clusters []*Cluster))

	RemoveCluster(cluster *Cluster, call func(err error, cluster *Cluster))

	SaveInstance(instance Instance, call func(err error, instance Instance))

	BatchUpdateInstance(instances []Instance, call func(err error, instances []Instance))

	RemoveInstance(instance Instance, call func(err error, instance Instance))
}

func InitDiscoveryStorage() (bool, error) {
	return plugin.RegisterPlugin(common.NewCtxPole(), &kvDiscoveryStorage{})
}

type kvDiscoveryStorage struct {
	hashRegionKv  *storage.HashRegionKVStorage
	rangeRegionKv *storage.RangeRegionKVStorage
}

func (kvD *kvDiscoveryStorage) Init(ctx *common.ContextPole) {

}

func (kvD *kvDiscoveryStorage) Run() {

}

func (kvD *kvDiscoveryStorage) SaveService(service *Service, call func(err error, service *Service)) {
	
}

func (kvD *kvDiscoveryStorage) BatchUpdateService(services []*Service, call func(err error, services []*Service)) {
	
}

func (kvD *kvDiscoveryStorage) RemoveService(services *Service, call func(err error, services *Service)) {
	
}

func (kvD *kvDiscoveryStorage) SaveCluster(cluster *Cluster, call func(err error, cluster *Cluster)) {
	
}

func (kvD *kvDiscoveryStorage) BatchUpdateCluster(clusters []*Cluster, call func(err error, clusters []*Cluster)) {
	
}

func (kvD *kvDiscoveryStorage) RemoveCluster(cluster *Cluster, call func(err error, cluster *Cluster)) {
	
}

func (kvD *kvDiscoveryStorage) SaveInstance(instance Instance, call func(err error, instance Instance)) {
	
}

func (kvD *kvDiscoveryStorage) BatchUpdateInstance(instances []Instance, call func(err error, instances []Instance)) {
	
}

func (kvD *kvDiscoveryStorage) RemoveInstance(instance Instance, call func(err error, instance Instance)) {
	
}

func (kvD *kvDiscoveryStorage) Name() string {
	return PluginDiscoveryStorage
}

func (kvD *kvDiscoveryStorage) Destroy() {

}
