package etcdsdk

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/tests/v3/integration"
)

type TestStruct struct {
	BaseInfo
	Foo string `json:"foo,omitempty"`
	Bar string `json:"bar"`
}

func (s *TestStruct) KeyPrefix() string {
	return "test_struct"
}

func etcdSetup(t *testing.T) *integration.ClusterV3 {
	integration.BeforeTestExternal(t)
	return integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
}

func TestGet(t *testing.T) {
	mockCluster := etcdSetup(t)
	defer mockCluster.Terminate(t)
	etcdClient := mockCluster.RandClient()

	tests := []struct {
		Desc       string
		giveKey    string
		giveType   reflect.Type
		givePrefix string
		formatFunc formatFunc
		wantErr    error
		obj        interface{}
	}{
		{
			Desc:     "get test - get prefix by KeyPrefix of the model",
			giveKey:  "test1",
			giveType: reflect.TypeOf(TestStruct{}),
			obj: &TestStruct{
				BaseInfo: BaseInfo{ID: "1"},
				Foo:      "f",
				Bar:      "b",
			},
		},
		{
			Desc:       "set cluster givePrefix",
			giveKey:    "test2",
			giveType:   reflect.TypeOf(TestStruct{}),
			givePrefix: "/test_prefix",
			obj: &TestStruct{
				BaseInfo: BaseInfo{ID: "1"},
				Foo:      "f",
				Bar:      "b",
			},
		},
		{
			Desc:     "not found",
			giveKey:  "test3",
			giveType: reflect.TypeOf(TestStruct{}),
			wantErr:  ErrNotFound,
		},
		{
			Desc:     "format function",
			giveKey:  "test4",
			giveType: reflect.TypeOf(TestStruct{}),
			formatFunc: func(key string, obj interface{}) interface{} {
				obj.(*TestStruct).Bar = ""
				return obj
			},
			obj: &TestStruct{
				BaseInfo: BaseInfo{ID: "1"},
				Foo:      "f",
				Bar:      "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Desc, func(t *testing.T) {
			sdk := New(etcdClient, nil, "/apisix")

			q := sdk.New().Type(tc.giveType)
			if tc.givePrefix != "" {
				q = q.Prefix(tc.givePrefix)
			}
			if tc.formatFunc != nil {
				q = q.Format(tc.formatFunc)
			}

			// create
			if tc.obj != nil {
				_, err := q.Create(context.Background(), tc.giveKey, tc.obj)
				assert.Nil(t, err)
			}

			r, err := q.Get(context.TODO(), tc.giveKey)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error(), "checking error of Get")
			}
			assert.Equal(t, tc.obj, r, "checking Result of Get")
		})
	}
}
