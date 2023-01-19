package etcdsdk

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// SDK is the interface for the SDK.
type SDK interface {
	// New creates a new query object
	New() Query
	// Close closes the SDK
	Close() error
}

// Query is the interface for all query objects
type Query interface {
	// Get returns the object of the given model type with the given Key.
	Get(ctx context.Context, key string) (interface{}, error)
	// List returns the list of objects of the given model type with the given Key.
	List(ctx context.Context) (*ListOutput, error)
	// Create creates the object of the given model type with the given Key.
	Create(ctx context.Context, key string, obj interface{}) (*clientv3.PutResponse, error)
	// Update updates the object of the given model type with the given Key.
	Update(ctx context.Context, key string, obj interface{}, createIfNotExist bool) (*clientv3.PutResponse, error)
	// Delete deletes the object of the given model type with the given Key.
	Delete(ctx context.Context, key string) (*clientv3.DeleteResponse, error)
	// Patch updates the object of the given model type with the given Key.
	Patch(ctx context.Context, key string, obj interface{}) (*clientv3.PutResponse, error)

	// Type sets the model type for the query.
	Type(typ reflect.Type) Query
	// Format sets the formatFunc function for the query.
	Format(format formatFunc) Query
	// Filter sets the filterFunc function for the query.
	Filter(filter filterFunc) Query
	// Sort sets the sortFunc function for the query.
	Sort(sort sortFunc) Query
	// Prefix sets the resource prefix for the query.
	Prefix(prefix string) Query
	// Page sets the page for the list function of the query.
	Page(page int) Query
	// PageSize sets the page size for the list function of the query.
	PageSize(pageSize int) Query
	// Hook registers the hook for the query.
	Hook(hook Hook) Query
	// GetResourcePrefix returns the key prefix of the resource.
	GetResourcePrefix() string
}

type sdk struct {
	client *clientv3.Client
	hooks  []Hook
	prefix string
}

type filterFunc func(key string, obj interface{}) bool
type formatFunc func(key string, obj interface{}) interface{}
type sortFunc func(i, j interface{}) bool

type query struct {
	filterFunc
	formatFunc
	sortFunc

	typ    reflect.Type
	client *clientv3.Client

	prefix         string
	resourcePrefix string

	page     int
	pageSize int

	hooks []Hook
}

// ListOutput is the output of the list function.
type ListOutput struct {
	Rows      []interface{} `json:"rows"`
	TotalSize int           `json:"total_size"`
}

// Prefixer is the interface that wraps the KeyPrefix Method
// to set Key resourcePrefix for specific model type.
type Prefixer interface {
	// KeyPrefix returns the Key prefix for the model type.
	KeyPrefix() string
}

// BaseInfo is the base info for most models.
type BaseInfo struct {
	ID         ID    `json:"id,omitempty"`
	CreateTime int64 `json:"create_time,omitempty"`
	UpdateTime int64 `json:"update_time,omitempty"`
}

// GetBaseInfo returns the base info of the model.
func (info *BaseInfo) GetBaseInfo() *BaseInfo {
	return info
}

// IBaseInfo is the interface that wraps the GetBaseInfo method
type IBaseInfo interface {
	GetBaseInfo() *BaseInfo
}

// ID is the type of the id field used for any entities
type ID string

// UnmarshalJSON implements the json.Unmarshaler interface for ID.
func (id *ID) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	switch v := value.(type) {
	case string:
		*id = ID(v)
	case float64:
		*id = ID(strconv.FormatUint(uint64(v), 10))
	default:
		panic("unknown type")
	}
	return nil
}

var DefaultSortFunc = func(i, j interface{}) bool {
	iBase := i.(IBaseInfo).GetBaseInfo()
	jBase := j.(IBaseInfo).GetBaseInfo()
	if iBase.UpdateTime != jBase.UpdateTime {
		return iBase.UpdateTime > jBase.UpdateTime
	}
	if iBase.CreateTime != jBase.CreateTime {
		return iBase.CreateTime > jBase.CreateTime
	}
	return iBase.ID < jBase.ID
}
