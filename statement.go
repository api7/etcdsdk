package etcdsdk

import (
	"context"
	"encoding/json"
	"path"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Type sets the model type for the query.
func (q *query) Type(typ reflect.Type) Query {
	q.typ = typ
	return q
}

// Format sets the formatFunc function for the query.
func (q *query) Format(format formatFunc) Query {
	q.formatFunc = format
	return q
}

// Filter sets the filterFunc function for the query.
func (q *query) Filter(filter filterFunc) Query {
	q.filterFunc = filter
	return q
}

// Hook registers the hook for the query.
func (q *query) Hook(hook Hook) Query {
	q.hooks = append(q.hooks, hook)
	return q
}

// Prefix sets the sortFunc function for the query.
func (q *query) Prefix(prefix string) Query {
	q.resourcePrefix = prefix
	return q
}

// Sort sets the sortFunc function for the query.
func (q *query) Sort(sort sortFunc) Query {
	q.sortFunc = sort
	return q
}

// Page sets the page for the list function of the query.
func (q *query) Page(page int) Query {
	q.page = page
	return q
}

// PageSize sets the page size for the list function of the query.
func (q *query) PageSize(pageSize int) Query {
	q.pageSize = pageSize
	return q
}

// stringToPtrObj converts the given string to an object of the special model.
func (q *query) stringToPtrObj(str string) (interface{}, error) {
	value := reflect.New(q.typ)
	ret := value.Interface()
	err := json.Unmarshal([]byte(str), ret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to json unmarshal")
	}

	return ret, nil
}

// realKey returns the real Key for the given Key.
func (q *query) realKey(_ context.Context, key string) string {
	resourcePrefix := q.GetResourcePrefix()
	key = path.Join(q.prefix, resourcePrefix, key)
	return key
}

// GetResourcePrefix gets the Key resource prefix for the given model type.
func (q *query) GetResourcePrefix() string {
	if q.resourcePrefix != "" {
		return q.resourcePrefix
	}

	value := reflect.New(q.typ)
	prefix := strings.ToLower(q.typ.Name())
	if prefixer, ok := value.Interface().(Prefixer); ok {
		prefix = prefixer.KeyPrefix()
	}
	return prefix
}
