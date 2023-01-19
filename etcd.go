package etcdsdk

import (
	"fmt"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	mutex     sync.Mutex
	clientMap = map[string]*etcdClientWrapper{}
)

type etcdClientWrapper struct {
	client     *clientv3.Client
	references int
}

func newEtcdClient(cfg clientv3.Config) (*clientv3.Client, error) {
	mutex.Lock()
	defer mutex.Unlock()
	clientMapKey := fmt.Sprintf("%v", cfg)
	if clientMap[clientMapKey] != nil {
		clientMap[clientMapKey].references++
		return clientMap[clientMapKey].client, nil
	}
	cli, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	clientMap[clientMapKey] = &etcdClientWrapper{
		client:     cli,
		references: 1,
	}
	return clientMap[clientMapKey].client, nil
}
