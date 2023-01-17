package etcdsdk

import (
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"path"
	"sort"

	jsonPatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
)

// New create a new query object
func (s *sdk) New() Query {
	return &query{
		client: s.client,
		hooks:  s.hooks,
		prefix: s.prefix,
	}
}

// Create creates the object of the given model type with the given Key.
func (q *query) Create(ctx context.Context, key string, obj interface{}) (*clientv3.PutResponse, error) {
	_, err := q.client.Get(ctx, q.realKey(ctx, key))
	if err == nil {
		return nil, ErrAlreadyExist
	}

	bs, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to json marshal")
	}
	resp, err := q.client.Put(ctx, q.realKey(ctx, key), string(bs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create")
	}

	q.runHooks(ctx, &HookParams{
		Method:   HookMethodCreate,
		Key:      key,
		Val:      obj,
		Revision: resp.Header.Revision,
		Result:   obj,
	})

	return resp, nil
}

// Update updates the object of the given model type with the given Key.
func (q *query) Update(ctx context.Context, key string, obj interface{}, createIfNotExist bool) (*clientv3.PutResponse, error) {
	// check if the Key exists
	_, err := q.Get(ctx, key)
	if err != nil {
		if err == ErrNotFound && createIfNotExist {
			return q.Create(ctx, key, obj)
		}
		return nil, errors.Wrap(err, "failed to get data")
	}
	//  update the Key
	bs, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to json marshal")
	}
	resp, err := q.client.Put(ctx, q.realKey(ctx, key), string(bs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to update")
	}

	q.runHooks(ctx, &HookParams{
		Method:   HookMethodUpdate,
		Key:      key,
		Val:      obj,
		Revision: resp.Header.Revision,
		Result:   obj,
	})

	return resp, nil
}

// Delete deletes data from db by the given Key.
func (q *query) Delete(ctx context.Context, key string) (*clientv3.DeleteResponse, error) {
	r, err := q.client.Get(ctx, q.realKey(ctx, key))
	if err != nil {
		return nil, err
	}
	val, _ := q.stringToPtrObj(string(r.Kvs[0].Value))

	resp, err := q.client.Delete(ctx, q.realKey(ctx, key))
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete")
	}

	q.runHooks(ctx, &HookParams{
		Method:   HookMethodDelete,
		Key:      key,
		Val:      val,
		Revision: resp.Header.Revision,
	})

	return resp, nil
}

// Get returns the object of the given model type with the given Key.
func (q *query) Get(ctx context.Context, key string) (interface{}, error) {
	resp, err := q.client.Get(ctx, q.realKey(ctx, key))
	if err != nil {
		return nil, err
	}

	val, err := q.stringToPtrObj(string(resp.Kvs[0].Value))
	if err != nil {
		return nil, errors.Wrap(err, "failed to bind string to struct object")
	}

	if q.formatFunc != nil {
		val = q.formatFunc(q.realKey(ctx, key), val)
	}

	q.runHooks(ctx, &HookParams{
		Method: HookMethodGet,
		Key:    key,
		Result: val,
	})

	return val, nil
}

// List returns the list of objects for the given model type of the query.
func (q *query) List(ctx context.Context) (*ListOutput, error) {
	keyPrefix := path.Join(q.prefix, q.GetResourcePrefix())
	resp, err := q.client.Get(ctx, keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	// filter and format
	ret := make([]interface{}, 0)
	for _, kv := range resp.Kvs {
		val, err := q.stringToPtrObj(string(kv.Value))
		if err != nil {
			return nil, errors.Wrap(err, "failed to bind string to struct object")
		}
		if q.filterFunc != nil && !q.filterFunc(string(kv.Key), val) {
			continue
		}
		if q.formatFunc != nil {
			val = q.formatFunc(string(kv.Key), val)
		}
		ret = append(ret, val)
	}

	// sort and page
	output := &ListOutput{
		TotalSize: len(ret),
	}
	if q.sortFunc == nil {
		q.sortFunc = DefaultSortFunc
	}
	sort.Slice(ret, func(i, j int) bool {
		return q.sortFunc(ret[i], ret[j])
	})
	output.Rows = Pagination(ret, q.pageSize, q.page)

	q.runHooks(ctx, &HookParams{
		Method: HookMethodList,
		Result: output,
	})

	return output, nil
}

// Patch updates the object of the given model type with the given Key.
func (q *query) Patch(ctx context.Context, key string, obj interface{}) (*clientv3.PutResponse, error) {
	r, err := q.client.Get(ctx, q.realKey(ctx, key))
	if err != nil {
		return nil, err
	}

	patch, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal")
	}

	result, err := jsonPatch.MergePatch(r.Kvs[0].Value, patch)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply patch")
	}

	resp, err := q.client.Put(ctx, q.realKey(ctx, key), string(result))
	if err != nil {
		return nil, errors.Wrap(err, "failed to update")
	}

	val, err := q.stringToPtrObj(string(result))
	if err != nil {
		return nil, errors.Wrap(err, "failed to bind string to struct object")
	}

	q.runHooks(ctx, &HookParams{
		Method:   HookMethodPatch,
		Key:      key,
		Val:      obj,
		Revision: r.Header.Revision,
		Result:   val,
	})

	return resp, nil
}
