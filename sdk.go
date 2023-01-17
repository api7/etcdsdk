package etcdsdk

import clientv3 "go.etcd.io/etcd/client/v3"

// New create sdk object
func New(config clientv3.Config, hooks []Hook, prefix string) (*sdk, error) {
	client, err := newEtcdClient(config)
	if err != nil {
		return nil, err
	}

	return &sdk{
		client: client,
		hooks:  hooks,
		prefix: prefix,
	}, nil
}

func (s *sdk) Close() error {
	return s.client.Close()
}
