package etcdsdk

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/errgo.v2/fmt/errors"

	"github.com/api7/dashboard/internal/pkg/consts"
	"github.com/api7/dashboard/pkg/db"
)

type Foo struct {
	Bar string `json:"bar"`
}

func TestDefaultHook(t *testing.T) {
	// skip hook due to Method not matched
	called := false
	var defaultHook = Hook{
		Name: "default",
		Methods: []consts.HookMethod{
			consts.HookMethodPatch,
			consts.HookMethodCreate,
			consts.HookMethodUpdate,
			consts.HookMethodDelete,
		},
		Handler: func(ctx context.Context, q Query, params *HookParams) {
			called = true
		},
	}
	ctrl := gomock.NewController(t)
	d := db.NewMockDB(ctrl)
	d.EXPECT().Get(gomock.Any(), "/management/foo/test1").Return("", db.ErrNotFound)
	svc := &sdk{
		db:            d,
		hooks:         []Hook{defaultHook},
		clusterPrefix: "/management",
	}
	_, _ = svc.New().Type(reflect.TypeOf(Foo{})).Get(context.TODO(), "test1")
	assert.False(t, called)

	// skip hook due to db error
	called = false
	d.EXPECT().Get(gomock.Any(), "/management/foo/test1").Return("", db.ErrNotFound)
	d.EXPECT().Create(gomock.Any(), "/management/foo/test1", `{"bar":"baz"}`).Return(int64(0), errors.New("db error"))
	svc = &sdk{
		db:            d,
		hooks:         []Hook{defaultHook},
		clusterPrefix: "/management",
	}
	_, _ = svc.New().Type(reflect.TypeOf(Foo{})).Create(context.TODO(), "test1", &Foo{Bar: "baz"})
	assert.False(t, called)

	// hook called
	called = false
	d.EXPECT().Get(gomock.Any(), "/management/foo/test1").Return("", db.ErrNotFound)
	d.EXPECT().Create(gomock.Any(), "/management/foo/test1", `{"bar":"baz"}`).Return(int64(1), nil)
	svc = &sdk{
		db:            d,
		hooks:         []Hook{defaultHook},
		clusterPrefix: "/management",
	}
	_, _ = svc.New().Type(reflect.TypeOf(Foo{})).Create(context.TODO(), "test1", &Foo{Bar: "baz"})
	assert.True(t, called)

	// hook called for Method all
	defaultHook.Methods = []consts.HookMethod{consts.HookMethodAll}
	called = false
	d.EXPECT().Get(gomock.Any(), "/management/foo/test1").Return("", db.ErrNotFound)
	d.EXPECT().Create(gomock.Any(), "/management/foo/test1", `{"bar":"baz"}`).Return(int64(1), nil)
	svc = &sdk{
		db:            d,
		hooks:         []Hook{defaultHook},
		clusterPrefix: "/management",
	}
	_, _ = svc.New().Type(reflect.TypeOf(Foo{})).Create(context.TODO(), "test1", &Foo{Bar: "baz"})
	assert.True(t, called)
}

func TestQueryHook(t *testing.T) {
	// skip hook due to Method not matched
	called := false
	hook := Hook{
		Methods: []consts.HookMethod{
			consts.HookMethodPatch,
		},
		Handler: func(ctx context.Context, q Query, params *HookParams) {
			called = true
		},
	}
	ctrl := gomock.NewController(t)
	d := db.NewMockDB(ctrl)
	d.EXPECT().Get(gomock.Any(), "/management/foo/test1").Return(`{"bar":"baz"}`, nil).AnyTimes()
	svc := &sdk{
		db:            d,
		clusterPrefix: "/management",
	}
	_, _ = svc.New().Type(reflect.TypeOf(Foo{})).Hook(hook).Get(context.TODO(), "test1")
	assert.False(t, called)

	// skip hook due to db error
	hook.Methods = nil
	called = false
	d = db.NewMockDB(ctrl)
	d.EXPECT().Get(gomock.Any(), "/management/foo/test1").Return("", errors.New("not found")).AnyTimes()
	svc = &sdk{
		db:            d,
		clusterPrefix: "/management",
	}
	_, _ = svc.New().Type(reflect.TypeOf(Foo{})).Hook(hook).Get(context.TODO(), "test1")
	assert.False(t, called)

	// hook called due to Method matched
	hook.Methods = []consts.HookMethod{
		consts.HookMethodGet,
		consts.HookMethodUpdate,
	}
	called = false
	d = db.NewMockDB(ctrl)
	d.EXPECT().Get(gomock.Any(), "/management/foo/test1").Return(`{"bar":"baz"}`, nil).AnyTimes()
	svc = &sdk{
		db:            d,
		clusterPrefix: "/management",
	}
	_, _ = svc.New().Type(reflect.TypeOf(Foo{})).Hook(hook).Get(context.TODO(), "test1")
	assert.True(t, called)

	// hook called due to Method all
	hook.Methods = []consts.HookMethod{
		consts.HookMethodAll,
	}
	called = false
	_, _ = svc.New().Type(reflect.TypeOf(Foo{})).Hook(hook).Get(context.TODO(), "test1")
	assert.True(t, called)
}
