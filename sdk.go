package etcdsdk

import clientv3 "go.etcd.io/etcd/client/v3"

// New create sdk object
func New(client *clientv3.Client, hooks []Hook, prefix string) *sdk {
	return &sdk{
		client: client,
		hooks:  hooks,
		prefix: prefix,
	}
}

func (s *sdk) Close() error {
	return s.client.Close()
}
