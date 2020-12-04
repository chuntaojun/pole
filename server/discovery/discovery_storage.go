package discovery

import (
	"sync"

	"github.com/Conf-Group/pole/pojo"
)

type DiscoveryStorage interface {
	SaveService(service *pojo.Service)

	BatchUpdateService(services []*pojo.Service)

	RemoveService(services *pojo.Service)

	SaveCluster(cluster *pojo.Cluster)

	BatchUpdateCluster(clusters []*pojo.Cluster)

	RemoveCluster(cluster *pojo.Cluster)

	SaveInstance(instance *pojo.Instance)

	BatchUpdateInstance(instances []*pojo.Instance)

	RemoveInstance(instance *pojo.Instance)
}



type memoryDiscoveryStorage struct {
	serviceMapLock  sync.RWMutex
	clusterMapLock  sync.RWMutex
	instanceMapLock sync.RWMutex
	services        map[string]*pojo.Service
}
