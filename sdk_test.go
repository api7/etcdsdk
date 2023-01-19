package etcdsdk

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	conf1   = clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}}
	prefix1 = "/test1"
	conf2   = clientv3.Config{Endpoints: []string{"127.0.0.2:2379"}}
	prefix2 = "/test2"
)

func TestSDKNewAndClose(t *testing.T) {
	// new etcd db
	d1, err := New(conf1, nil, prefix1)
	assert.Nil(t, err, "checking new etcd db")
	// new etcd db with same config and different prefix
	d2, err := New(conf1, nil, prefix2)
	assert.Nil(t, err, "checking new etcd db")
	// new etcd db with another config and same prefix with the first one
	d3, err := New(conf2, nil, prefix1)
	assert.Nil(t, err, "checking new etcd db")
	// close etcd db
	err = d1.Close()
	assert.Nil(t, err, "checking etcd db close")
	clientMapKey1 := fmt.Sprintf("%v", conf1)
	assert.NotNil(t, clientMap[clientMapKey1], "checking etcd db close")
	assert.Equal(t, 1, clientMap[clientMapKey1].references, "checking etcd db references")
	// close etcd db
	err = d2.Close()
	assert.Nil(t, err, "checking etcd db close")
	assert.Nil(t, clientMap[clientMapKey1], "checking etcd db close")
	// close etcd db
	err = d3.Close()
	assert.Nil(t, err, "checking etcd db close")
	clientMapKey2 := fmt.Sprintf("%v", conf2)
	assert.Nil(t, clientMap[clientMapKey2], "checking etcd db close")
}
