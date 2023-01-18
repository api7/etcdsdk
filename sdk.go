package etcdsdk

import (
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// New create sdk object
func New(config clientv3.Config, hooks []Hook, prefix string) (SDK, error) {
	client, err := newEtcdClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create etcd client")
	}

	return &sdk{
		client: client,
		hooks:  hooks,
		prefix: prefix,
	}, nil
}

func (s *sdk) Close() error {
	mutex.Lock()
	defer mutex.Unlock()
	cli := s.client
	for k, clientWrapper := range clientMap {
		if clientWrapper.client == cli {
			clientWrapper.references--
			if clientWrapper.references == 0 {
				delete(clientMap, k)
				return cli.Close()
			}
		}
	}

	return nil
}
