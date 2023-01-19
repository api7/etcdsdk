package etcdsdk

import "errors"

// HookMethod defines which method the hook will be triggered
type HookMethod string

const (
	// HookMethodAll means the hook will be triggered on all methods
	HookMethodAll HookMethod = "all"
	// HookMethodGet means the hook will be triggered on Get method
	HookMethodGet HookMethod = "get"
	// HookMethodList means the hook will be triggered on List method
	HookMethodList HookMethod = "list"
	// HookMethodCreate means the hook will be triggered on Create method
	HookMethodCreate HookMethod = "create"
	// HookMethodUpdate means the hook will be triggered on Update method
	HookMethodUpdate HookMethod = "update"
	// HookMethodDelete means the hook will be triggered on Delete method
	HookMethodDelete HookMethod = "delete"
	// HookMethodPatch means the hook will be triggered on Patch method
	HookMethodPatch HookMethod = "patch"
)

var (
	// ErrNotFound means target was not found.
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExist means target already exists.
	ErrAlreadyExist = errors.New("already exists")
)
