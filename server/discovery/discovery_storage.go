package discovery

import (
	"sync"

	"github.com/Conf-Group/pole/pojo"
)

type DiscoveryStorage interface {
	SaveService(service Service)

	BatchUpdateService(services []Service)

	RemoveService(services Service)

	SaveCluster(cluster Cluster)

	BatchUpdateCluster(clusters Cluster)

	RemoveCluster(cluster Cluster)

	SaveInstance(instance Instance)

	BatchUpdateInstance(instances Instance)

	RemoveInstance(instance Instance)
}

type memoryDiscoveryStorage struct {
	serviceMapLock  sync.RWMutex
	clusterMapLock  sync.RWMutex
	instanceMapLock sync.RWMutex
	services        map[string]*pojo.Service
}

type pebbleDiscoveryStorage struct {
}
