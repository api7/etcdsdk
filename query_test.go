package etcdsdk

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/tests/v3/integration"
)

type TestStruct struct {
	BaseInfo
	Foo string `json:"foo,omitempty"`
	Bar string `json:"bar,omitempty"`
}

func (s *TestStruct) KeyPrefix() string {
	return "test_struct"
}

func etcdSetup(t *testing.T) *integration.ClusterV3 {
	integration.BeforeTestExternal(t)
	return integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
}

func TestCreateAndGet(t *testing.T) {
	mockCluster := etcdSetup(t)
	defer mockCluster.Terminate(t)
	var endpoints []string
	for _, member := range mockCluster.Members {
		endpoints = append(endpoints, member.GRPCURL())
	}

	tests := []struct {
		Desc       string
		giveKey    string
		giveType   reflect.Type
		givePrefix string
		formatFunc formatFunc
		createErr  error
		getErr     error
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
			Desc:       "set cluster giveResourcePrefix",
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
			Desc:       "create failed (object already exist)",
			giveKey:    "test2",
			giveType:   reflect.TypeOf(TestStruct{}),
			givePrefix: "/test_prefix",
			obj: &TestStruct{
				BaseInfo: BaseInfo{ID: "1"},
				Foo:      "f",
				Bar:      "b",
			},
			createErr: ErrAlreadyExist,
		},
		{
			Desc:     "not found",
			giveKey:  "test3",
			giveType: reflect.TypeOf(TestStruct{}),
			getErr:   ErrNotFound,
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
			sdk, err := New(clientv3.Config{Endpoints: endpoints}, nil, "/apisix")
			assert.Nil(t, err)

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
				if tc.createErr != nil {
					assert.Equal(t, tc.createErr, err)
				} else {
					assert.Nil(t, err)
				}
			}

			r, err := q.Get(context.TODO(), tc.giveKey)
			if tc.getErr != nil {
				assert.Equal(t, tc.getErr.Error(), err.Error(), "checking error of Get")
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tc.obj, r, "checking Result of Get")
		})
	}
}

func TestList(t *testing.T) {
	mockCluster := etcdSetup(t)
	defer mockCluster.Terminate(t)
	var endpoints []string
	for _, member := range mockCluster.Members {
		endpoints = append(endpoints, member.GRPCURL())
	}

	mockData := map[string]string{
		"/apisix/test_prefix/1": `{"id":1,"create_time":11,"update_time":111,"foo":"f","bar":"b"}`,
		"/apisix/test_prefix/2": `{"id":2,"create_time":22,"update_time":22,"foo":"f2","bar":"b2"}`,
		"/apisix/test_prefix/3": `{"id":3,"create_time":33,"update_time":333,"foo":"f3","bar":"b3"}`,
		"/api7/test_struct/1":   `{"id":1,"create_time":11,"update_time":111,"foo":"f","bar":"b"}`,
		"/api7/test_struct/2":   `{"id":2,"create_time":22,"update_time":22,"foo":"f2","bar":"b2"}`,
		"/api7/test_struct/3":   `{"id":3,"create_time":33,"update_time":333,"foo":"f3","bar":"b3"}`,
		"/apisix/string/1":      `this is a string`,
	}
	for k, v := range mockData {
		mockCluster.RandClient().Put(context.Background(), k, v)
	}

	tests := []struct {
		Desc               string
		giveType           reflect.Type
		givePrefix         string
		giveResourcePrefix string
		formatFunc         formatFunc
		filterFunc         filterFunc
		sortFunc           sortFunc
		page               int
		pageSize           int
		wantErr            error
		wantResult         *ListOutput
	}{
		{
			Desc:               "list test (no array object)",
			giveType:           reflect.TypeOf(TestStruct{}),
			givePrefix:         "/apisix",
			giveResourcePrefix: "/not_exist",
			wantResult: &ListOutput{
				Rows:      []interface{}{},
				TotalSize: 0,
			},
		},
		{
			Desc:       "list test",
			giveType:   reflect.TypeOf(TestStruct{}),
			givePrefix: "/api7",
			wantResult: &ListOutput{
				Rows: []interface{}{
					&TestStruct{
						BaseInfo: BaseInfo{
							ID:         "3",
							CreateTime: 33,
							UpdateTime: 333,
						},
						Foo: "f3",
						Bar: "b3",
					},
					&TestStruct{
						BaseInfo: BaseInfo{
							ID:         "1",
							CreateTime: 11,
							UpdateTime: 111,
						},
						Foo: "f",
						Bar: "b",
					},
					&TestStruct{
						BaseInfo: BaseInfo{
							ID:         "2",
							CreateTime: 22,
							UpdateTime: 22,
						},
						Foo: "f2",
						Bar: "b2",
					},
				},
				TotalSize: 3,
			},
		},
		{
			Desc:       "filter test",
			giveType:   reflect.TypeOf(TestStruct{}),
			givePrefix: "/api7",
			filterFunc: func(key string, obj interface{}) bool {
				return obj.(*TestStruct).Foo == "f2"
			},
			wantResult: &ListOutput{
				Rows: []interface{}{
					&TestStruct{
						BaseInfo: BaseInfo{
							ID:         "2",
							CreateTime: 22,
							UpdateTime: 22,
						},
						Foo: "f2",
						Bar: "b2",
					},
				},
				TotalSize: 1,
			},
		},
		{
			Desc:               "sort and page test",
			giveType:           reflect.TypeOf(TestStruct{}),
			givePrefix:         "/apisix",
			giveResourcePrefix: "/test_prefix",
			sortFunc: func(a interface{}, b interface{}) bool {
				return a.(*TestStruct).ID > b.(*TestStruct).ID
			},
			page:     1,
			pageSize: 1,
			wantResult: &ListOutput{
				Rows: []interface{}{
					&TestStruct{
						BaseInfo: BaseInfo{
							ID:         "3",
							CreateTime: 33,
							UpdateTime: 333,
						},
						Foo: "f3",
						Bar: "b3",
					},
				},
				TotalSize: 3,
			},
		},
		{
			Desc:               "invalid data",
			giveType:           reflect.TypeOf(TestStruct{}),
			givePrefix:         "/apisix",
			giveResourcePrefix: "/string",
			wantErr:            errors.New("failed to bind string to struct object: failed to json unmarshal: invalid character 'h' in literal true (expecting 'r')"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			sdk, err := New(clientv3.Config{Endpoints: endpoints}, nil, tc.givePrefix)
			assert.Nil(t, err)

			q := sdk.New().Type(tc.giveType)
			if tc.giveResourcePrefix != "" {
				q = q.Prefix(tc.giveResourcePrefix)
			}
			if tc.formatFunc != nil {
				q = q.Format(tc.formatFunc)
			}
			if tc.filterFunc != nil {
				q = q.Filter(tc.filterFunc)
			}
			if tc.sortFunc != nil {
				q = q.Sort(tc.sortFunc)
			}
			if tc.page != 0 {
				q = q.Page(tc.page)
			}
			if tc.pageSize != 0 {
				q = q.PageSize(tc.pageSize)
			}
			r, err := q.List(context.TODO())
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error(), "checking error of List")
			}
			assert.Equal(t, tc.wantResult, r, "checking Result of List")
		})
	}
}

func TestUpdate(t *testing.T) {
	mockCluster := etcdSetup(t)
	defer mockCluster.Terminate(t)
	var endpoints []string
	for _, member := range mockCluster.Members {
		endpoints = append(endpoints, member.GRPCURL())
	}

	mockData := map[string]string{
		"/api7/test_struct/1": `{"id":1,"create_time":11,"update_time":111,"foo":"f","bar":"b"}`,
	}
	for k, v := range mockData {
		mockCluster.RandClient().Put(context.Background(), k, v)
	}

	tests := []struct {
		Desc               string
		givePrefix         string
		giveKey            string
		giveType           reflect.Type
		giveValue          interface{}
		giveResourcePrefix string
		createNotExist     bool
		wantErr            error
	}{
		{
			Desc:       "update test",
			giveKey:    "1",
			givePrefix: "/api7",
			giveType:   reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: BaseInfo{
					ID:         "11",
					CreateTime: 11,
					UpdateTime: 11,
				},
				Foo: "f1",
				Bar: "b1",
			},
		},
		{
			Desc:       "not exits and failed test",
			giveKey:    "2",
			givePrefix: "/api7",
			giveType:   reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			createNotExist:     false,
			giveResourcePrefix: "/prefix1",
			wantErr:            errors.New("failed to get data: not found"),
		},
		{
			Desc:       "not exits and create test",
			giveKey:    "2",
			givePrefix: "/api7",
			giveType:   reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			createNotExist:     true,
			giveResourcePrefix: "/prefix1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Desc, func(t *testing.T) {
			sdk, err := New(clientv3.Config{Endpoints: endpoints}, nil, tc.givePrefix)
			assert.Nil(t, err)

			q := sdk.New().Type(tc.giveType)

			if tc.giveResourcePrefix != "" {
				q = q.Prefix(tc.giveResourcePrefix)
			}
			resp, err := q.Update(context.TODO(), tc.giveKey, tc.giveValue, tc.createNotExist)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error(), "checking error of Update")
			} else {
				assert.Nil(t, err, "checking error of Update")
				assert.Greater(t, resp.Header.Revision, int64(0), "checking revision of Update")

				// get the updated data from etcd
				resp, err := q.Get(context.TODO(), tc.giveKey)
				assert.Nil(t, err, "checking error of Get")
				assert.Equal(t, tc.giveValue, resp, "checking value of Get")
			}
		})
	}
}

func TestDelete(t *testing.T) {
	mockCluster := etcdSetup(t)
	defer mockCluster.Terminate(t)
	var endpoints []string
	for _, member := range mockCluster.Members {
		endpoints = append(endpoints, member.GRPCURL())
	}

	mockData := map[string]string{
		"/api7/test_struct/1": `{"id":1,"create_time":11,"update_time":111,"foo":"f","bar":"b"}`,
		"/apisix/prefix1/1":   `{"id":1,"create_time":11,"update_time":111,"foo":"f","bar":"b"}`,
	}
	for k, v := range mockData {
		mockCluster.RandClient().Put(context.Background(), k, v)
	}

	tests := []struct {
		Desc               string
		giveKey            string
		givePrefix         string
		giveType           reflect.Type
		giveResourcePrefix string
		wantErr            error
	}{
		{
			Desc:       "delete success",
			givePrefix: "/api7",
			giveKey:    "1",
			giveType:   reflect.TypeOf(TestStruct{}),
		},
		{
			Desc:       "delete failed (object not found)",
			giveKey:    "2",
			givePrefix: "/api7",
			giveType:   reflect.TypeOf(TestStruct{}),
			wantErr:    ErrNotFound,
		},
		{
			Desc:               "prefix test",
			giveKey:            "1",
			givePrefix:         "/apisix",
			giveType:           reflect.TypeOf(TestStruct{}),
			giveResourcePrefix: "/prefix1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Desc, func(t *testing.T) {
			sdk, err := New(clientv3.Config{Endpoints: endpoints}, nil, tc.givePrefix)
			assert.Nil(t, err)

			q := sdk.New().Type(tc.giveType)

			if tc.giveResourcePrefix != "" {
				q = q.Prefix(tc.giveResourcePrefix)
			}
			resp, err := q.Delete(context.TODO(), tc.giveKey)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error(), "checking error of Delete")
			} else {
				assert.Nil(t, err, "checking error of Delete")
				assert.Greater(t, resp.Header.Revision, int64(0), "checking revision of Update")

				// try to get the deleted data from etcd
				_, err := q.Get(context.TODO(), tc.giveKey)
				assert.Equal(t, ErrNotFound.Error(), err.Error(), "checking error of Get")
			}
		})
	}
}

func TestPatch(t *testing.T) {
	mockCluster := etcdSetup(t)
	defer mockCluster.Terminate(t)
	var endpoints []string
	for _, member := range mockCluster.Members {
		endpoints = append(endpoints, member.GRPCURL())
	}

	mockData := map[string]string{
		"/api7/test_struct/1": `{"id":1,"create_time":11,"update_time":111,"foo":"f","bar":"b"}`,
		"/apisix/prefix1/1":   `{"id":1,"create_time":11,"update_time":111,"foo":"f1","bar":"b1"}`,
		"/apisix/prefix1/2":   `this is a string`,
	}
	for k, v := range mockData {
		mockCluster.RandClient().Put(context.Background(), k, v)
	}

	tests := []struct {
		Desc               string
		giveKey            string
		giveType           reflect.Type
		givePrefix         string
		giveValue          interface{}
		expectValue        interface{}
		giveResourcePrefix string
		createNotExist     bool
		wantErr            error
	}{
		{
			Desc:       "patch object failed (not found)",
			giveKey:    "2",
			givePrefix: "/api7",
			giveType:   reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			wantErr: ErrNotFound,
		},
		{
			Desc:               "patch object failed (original data is not a json)",
			giveKey:            "2",
			givePrefix:         "/apisix",
			giveType:           reflect.TypeOf(TestStruct{}),
			giveResourcePrefix: "/prefix1",
			giveValue: &TestStruct{
				BaseInfo: BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			wantErr: errors.New("failed to bind string to struct object: failed to json unmarshal: invalid character 'h' in literal true (expecting 'r')"),
		},
		{
			Desc:               "patch object failed (patch data is not a json)",
			giveKey:            "1",
			givePrefix:         "/apisix",
			giveType:           reflect.TypeOf(TestStruct{}),
			giveResourcePrefix: "/prefix1",
			giveValue:          `this is a string`,
			wantErr:            errors.New("failed to apply patch: Invalid JSON Patch"),
		},
		{
			Desc:       "patch object success",
			giveKey:    "1",
			givePrefix: "/api7",
			giveType:   reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: BaseInfo{
					CreateTime: 22,
					UpdateTime: 22,
				},
				Bar: "b2",
			},
			expectValue: &TestStruct{
				BaseInfo: BaseInfo{
					ID:         "1",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f",
				Bar: "b2",
			},
		},
		{
			Desc:               "patch object success(give resource prefix)",
			giveKey:            "1",
			givePrefix:         "/apisix",
			giveResourcePrefix: "/prefix1",
			giveType:           reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: BaseInfo{
					CreateTime: 22,
					UpdateTime: 22,
				},
				Bar: "b2",
			},
			expectValue: &TestStruct{
				BaseInfo: BaseInfo{
					ID:         "1",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f1",
				Bar: "b2",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			sdk, err := New(clientv3.Config{Endpoints: endpoints}, nil, tc.givePrefix)
			assert.Nil(t, err)

			q := sdk.New().Type(tc.giveType)

			if tc.giveResourcePrefix != "" {
				q = q.Prefix(tc.giveResourcePrefix)
			}

			resp, err := q.Patch(context.TODO(), tc.giveKey, tc.giveValue)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error(), "checking error of Update")
				assert.Nil(t, resp, "checking Result of Update")
			} else {
				assert.Nil(t, err, "checking error of Update")
				assert.Greater(t, resp.Header.Revision, int64(0), "checking revision of Update")

				// get the updated data from etcd
				resp, err := q.Get(context.TODO(), tc.giveKey)
				assert.Nil(t, err, "checking error of Get")
				assert.Equal(t, tc.expectValue, resp, "checking value of Get")
			}
		})
	}
}
