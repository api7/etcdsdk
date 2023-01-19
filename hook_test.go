package etcdsdk

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Foo struct {
	Bar string `json:"bar"`
}

func TestDefaultHook(t *testing.T) {
	mockCluster := etcdSetup(t)
	defer mockCluster.Terminate(t)
	var endpoints []string
	for _, member := range mockCluster.Members {
		endpoints = append(endpoints, member.GRPCURL())
	}

	// skip hook due to Method not matched
	called := false
	var defaultHook = Hook{
		Name: "default",
		Methods: []HookMethod{
			HookMethodPatch,
			HookMethodCreate,
			HookMethodUpdate,
			HookMethodDelete,
		},
		Handler: func(ctx context.Context, q Query, params *HookParams) {
			called = true
		},
	}
	sdk, err := New(clientv3.Config{Endpoints: endpoints}, []Hook{defaultHook}, "/apisix")
	assert.Nil(t, err)

	_, _ = sdk.New().Type(reflect.TypeOf(Foo{})).Get(context.TODO(), "test1")
	assert.False(t, called)

	// skip hook due to error
	called = false
	_, _ = sdk.New().Type(reflect.TypeOf(Foo{})).Update(context.TODO(), "test1", &Foo{Bar: "baz"}, false)
	assert.False(t, called)

	// hook called
	called = false
	_, _ = sdk.New().Type(reflect.TypeOf(Foo{})).Create(context.TODO(), "test1", &Foo{Bar: "baz"})
	assert.True(t, called)

	// hook called for Method all
	defaultHook.Methods = []HookMethod{HookMethodAll}
	sdk, err = New(clientv3.Config{Endpoints: endpoints}, []Hook{defaultHook}, "/apisix")
	assert.Nil(t, err)

	called = false
	_, _ = sdk.New().Type(reflect.TypeOf(Foo{})).Create(context.TODO(), "test1", &Foo{Bar: "baz"})
	assert.True(t, called)
}

func TestQueryHook(t *testing.T) {
	mockCluster := etcdSetup(t)
	defer mockCluster.Terminate(t)
	var endpoints []string
	for _, member := range mockCluster.Members {
		endpoints = append(endpoints, member.GRPCURL())
	}

	// skip hook due to Method not matched
	called := false
	hook := Hook{
		Methods: []HookMethod{
			HookMethodPatch,
		},
		Handler: func(ctx context.Context, q Query, params *HookParams) {
			called = true
		},
	}
	sdk, err := New(clientv3.Config{Endpoints: endpoints}, nil, "/apisix")
	assert.Nil(t, err)

	_, _ = sdk.New().Type(reflect.TypeOf(Foo{})).Hook(hook).Get(context.TODO(), "test1")
	assert.False(t, called)

	// skip hook due to error
	hook.Methods = nil
	called = false
	_, err = sdk.New().Type(reflect.TypeOf(Foo{})).Hook(hook).Get(context.TODO(), "test1")
	assert.False(t, called)
	assert.Equal(t, "not found", err.Error())

	// hook called due to Method matched
	hook.Methods = []HookMethod{
		HookMethodGet,
		HookMethodUpdate,
		HookMethodCreate,
	}
	called = false
	_, err = sdk.New().Type(reflect.TypeOf(Foo{})).Hook(hook).Create(context.TODO(), "test1", &Foo{Bar: "baz"})
	assert.Nil(t, err)
	assert.True(t, called)

	called = false
	_, err = sdk.New().Type(reflect.TypeOf(Foo{})).Hook(hook).Get(context.TODO(), "test1")
	assert.Nil(t, err)
	assert.True(t, called)

	// not called due to no hook
	called = false
	_, err = sdk.New().Type(reflect.TypeOf(Foo{})).Create(context.TODO(), "test2", &Foo{Bar: "baz"})
	assert.Nil(t, err)
	assert.False(t, called)

	// hook called due to Method all
	hook.Methods = []HookMethod{
		HookMethodAll,
	}
	called = false
	_, err = sdk.New().Type(reflect.TypeOf(Foo{})).Hook(hook).Get(context.TODO(), "test1")
	assert.Nil(t, err)
	assert.True(t, called)
}
