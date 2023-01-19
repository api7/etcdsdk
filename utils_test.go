package etcdsdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInArray(t *testing.T) {
	stringTests := []struct {
		name string
		key  string
		arr  []string
		want bool
	}{
		{
			name: "test nil array",
			key:  "a",
			arr:  nil,
			want: false,
		},
		{
			name: "test empty array",
			key:  "a",
			arr:  []string{},
			want: false,
		},
		{
			name: "test not in array",
			key:  "d",
			arr:  []string{"a", "b", "c"},
			want: false,
		},
		{
			name: "test in array",
			key:  "a",
			arr:  []string{"a", "b", "c"},
			want: true,
		},
	}
	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			got := InArray(tt.key, tt.arr)
			assert.Equal(t, tt.want, got)
		})
	}

	methodTests := []struct {
		name string
		key  HookMethod
		arr  []HookMethod
		want bool
	}{
		{
			name: "test nil array",
			key:  HookMethodGet,
			arr:  nil,
			want: false,
		},
		{
			name: "test empty array",
			key:  HookMethodGet,
			arr:  []HookMethod{},
			want: false,
		},
		{
			name: "test not in array",
			key:  HookMethodDelete,
			arr:  []HookMethod{HookMethodGet, HookMethodList, HookMethodCreate},
			want: false,
		},
		{
			name: "test in array",
			key:  HookMethodPatch,
			arr:  []HookMethod{HookMethodGet, HookMethodUpdate, HookMethodCreate, HookMethodPatch},
			want: true,
		},
	}
	for _, tt := range methodTests {
		t.Run(tt.name, func(t *testing.T) {
			got := InArray(tt.key, tt.arr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPagination(t *testing.T) {
	tests := []struct {
		desc   string
		rows   []interface{}
		size   int
		page   int
		wanted []interface{}
	}{
		{
			desc:   "test nil array",
			rows:   nil,
			size:   10,
			page:   1,
			wanted: nil,
		},
		{
			desc:   "test empty array",
			rows:   []interface{}{},
			size:   10,
			page:   1,
			wanted: []interface{}{},
		},
		{
			desc:   "test all in one page",
			rows:   []interface{}{"a", "b", "c"},
			size:   10,
			page:   1,
			wanted: []interface{}{"a", "b", "c"},
		},
		{
			desc:   "test page size is 0",
			rows:   []interface{}{"a", "b", "c"},
			size:   0,
			page:   1,
			wanted: []interface{}{"a", "b", "c"},
		},
		{
			desc:   "test page is 0",
			rows:   []interface{}{"a", "b", "c"},
			size:   1,
			page:   0,
			wanted: []interface{}{"a", "b", "c"},
		},
		{
			desc:   "test first page",
			rows:   []interface{}{"a", "b", "c"},
			size:   2,
			page:   1,
			wanted: []interface{}{"a", "b"},
		},
		{
			desc:   "test last page",
			rows:   []interface{}{"a", "b", "c"},
			size:   2,
			page:   2,
			wanted: []interface{}{"c"},
		},
		{
			desc:   "test page exceeds",
			rows:   []interface{}{"a", "b", "c"},
			size:   2,
			page:   4,
			wanted: []interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := Pagination(tt.rows, tt.size, tt.page)
			assert.Equal(t, tt.wanted, got)
		})
	}
}
