package etcdsdk

import (
	"context"
)

// Hook is a hook function that will be called before and after the CRUD method.
type Hook struct {
	// Name is the name of the hook.
	Name string
	// Methods is the methods that the hook will be called.
	Methods []HookMethod
	// Handler is the hook function.
	Handler func(context.Context, Query, *HookParams)
}

// HookParams is the parameters passed to the hook handler.
type HookParams struct {
	// Method is the method to hook.
	Method HookMethod
	// Key is the parameter key of the method.
	Key string
	// Val is the parameter value of the method.
	Val interface{}
	// Revision is the etcd revision of the key.
	Revision int64
	// Result is the result of the method.
	Result interface{}
}

func (q *query) runHooks(ctx context.Context, params *HookParams) {
	for _, hook := range q.hooks {
		if InArray(HookMethodAll, hook.Methods) ||
			InArray(params.Method, hook.Methods) {
			hook.Handler(ctx, q, params)
		}
	}
}
