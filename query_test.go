package etcdsdk

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/api7/dashboard/pkg/db"
	"github.com/api7/dashboard/pkg/types"
)

type TestStruct struct {
	types.BaseInfo
	Foo string `json:"foo,omitempty"`
	Bar string `json:"bar"`
}

func (s *TestStruct) KeyPrefix() string {
	return "test_struct"
}

func TestGet(t *testing.T) {
	tests := []struct {
		Desc           string
		giveKey        string
		giveType       reflect.Type
		givePrefix     string
		formatFunc     formatFunc
		mockNewService func(*testing.T) *sdk
		wantErr        error
		wantResult     interface{}
	}{
		{
			Desc:     "get test - get prefix by KeyPrefix of the model",
			giveKey:  "test1",
			giveType: reflect.TypeOf(TestStruct{}),
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/test_struct/test1").
					Return(`{"id":1,"foo":"f","bar":"b"}`, nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantResult: &TestStruct{
				BaseInfo: types.BaseInfo{ID: "1"},
				Foo:      "f",
				Bar:      "b",
			},
		},
		{
			Desc:       "set cluster givePrefix",
			giveKey:    "test2",
			giveType:   reflect.TypeOf(TestStruct{}),
			givePrefix: "/test_prefix",
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/test_prefix/test2").
					Return(`{"id":1,"foo":"f","bar":"b"}`, nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantResult: &TestStruct{
				BaseInfo: types.BaseInfo{ID: "1"},
				Foo:      "f",
				Bar:      "b",
			},
		},
		{
			Desc:     "not found",
			giveKey:  "test3",
			giveType: reflect.TypeOf(TestStruct{}),
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/test_struct/test3").
					Return("", db.ErrNotFound)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantErr: db.ErrNotFound,
		},
		{
			Desc:     "format function",
			giveKey:  "test4",
			giveType: reflect.TypeOf(TestStruct{}),
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/test_struct/test4").
					Return(`{"id":"1","foo":"f","bar":"b"}`, nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			formatFunc: func(key string, obj interface{}) interface{} {
				obj.(*TestStruct).Bar = ""
				return obj
			},
			wantResult: &TestStruct{
				BaseInfo: types.BaseInfo{ID: "1"},
				Foo:      "f",
				Bar:      "",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			s := tc.mockNewService(t)
			q := s.New().Type(tc.giveType)
			if tc.givePrefix != "" {
				q = q.Prefix(tc.givePrefix)
			}
			if tc.formatFunc != nil {
				q = q.Format(tc.formatFunc)
			}
			r, err := q.Get(context.TODO(), tc.giveKey)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error(), "checking error of Get")
			}
			assert.Equal(t, tc.wantResult, r, "checking Result of Get")
		})
	}
}

func TestList(t *testing.T) {
	listRet := []db.KV{
		{
			Key:   "/management/test_prefix/1",
			Value: `{"id":1,"create_time":11,"update_time":111,"foo":"f","bar":"b"}`,
		},
		{
			Key:   "/management/test_prefix/2",
			Value: `{"id":2,"create_time":22,"update_time":22,"foo":"f2","bar":"b2"}`,
		},
		{
			Key:   "/management/test_prefix/3",
			Value: `{"id":3,"create_time":33,"update_time":333,"foo":"f3","bar":"b3"}`,
		},
	}

	tests := []struct {
		Desc           string
		giveType       reflect.Type
		givePrefix     string
		formatFunc     formatFunc
		filterFunc     filterFunc
		sortFunc       sortFunc
		page           int
		pageSize       int
		mockNewService func(*testing.T) *sdk
		wantErr        error
		wantResult     *ListOutput
	}{
		{
			Desc:     "list test (no array object)",
			giveType: reflect.TypeOf(TestStruct{}),
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().List(gomock.Any(), "/management/test_struct").
					Return([]db.KV{}, nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantResult: &ListOutput{
				Rows:      []interface{}{},
				TotalSize: 0,
			},
		},
		{
			Desc:     "list test",
			giveType: reflect.TypeOf(TestStruct{}),
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().List(gomock.Any(), "/management/test_struct").
					Return(listRet, nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantResult: &ListOutput{
				Rows: []interface{}{
					&TestStruct{
						BaseInfo: types.BaseInfo{
							ID:         "3",
							CreateTime: 33,
							UpdateTime: 333,
						},
						Foo: "f3",
						Bar: "b3",
					},
					&TestStruct{
						BaseInfo: types.BaseInfo{
							ID:         "1",
							CreateTime: 11,
							UpdateTime: 111,
						},
						Foo: "f",
						Bar: "b",
					},
					&TestStruct{
						BaseInfo: types.BaseInfo{
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
			Desc:     "filter test",
			giveType: reflect.TypeOf(TestStruct{}),
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().List(gomock.Any(), "/management/test_struct").
					Return(listRet, nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			filterFunc: func(key string, obj interface{}) bool {
				return obj.(*TestStruct).Foo == "f2"
			},
			wantResult: &ListOutput{
				Rows: []interface{}{
					&TestStruct{
						BaseInfo: types.BaseInfo{
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
			Desc:     "sort and page test",
			giveType: reflect.TypeOf(TestStruct{}),
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().List(gomock.Any(), "/management/test_struct").
					Return(listRet, nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			sortFunc: func(a interface{}, b interface{}) bool {
				return a.(*TestStruct).ID > b.(*TestStruct).ID
			},
			page:     1,
			pageSize: 1,
			wantResult: &ListOutput{
				Rows: []interface{}{
					&TestStruct{
						BaseInfo: types.BaseInfo{
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
			Desc:     "db list err",
			giveType: reflect.TypeOf(TestStruct{}),
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().List(gomock.Any(), "/management/test_struct").
					Return(nil, errors.New("list error"))
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantErr: errors.New("list error"),
		},
		{
			Desc:       "resource prefix test",
			giveType:   reflect.TypeOf(TestStruct{}),
			givePrefix: "/test_prefix",
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().List(gomock.Any(), "/management/test_prefix").
					Return(nil, errors.New("list error"))
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantErr: errors.New("list error"),
		},
		{
			Desc:       "invalid data",
			giveType:   reflect.TypeOf(TestStruct{}),
			givePrefix: "/test_prefix",
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().List(gomock.Any(), "/management/test_prefix").
					Return([]db.KV{
						{
							Key:   "/management/test_prefix/1",
							Value: `str`,
						},
						{
							Key:   "/management/test_prefix/2",
							Value: `{"id":2,"create_time":22,"update_time":22,"foo":"f2","bar":"b2"}`,
						},
					}, nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantErr: errors.New("failed to bind string to struct object: failed to json unmarshal: invalid character 's' looking for beginning of value"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			s := tc.mockNewService(t)
			q := s.New().Type(tc.giveType)
			if tc.givePrefix != "" {
				q = q.Prefix(tc.givePrefix)
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

func TestCreate(t *testing.T) {
	tests := []struct {
		Desc           string
		giveKey        string
		giveType       reflect.Type
		giveValue      interface{}
		givePrefix     string
		mockNewService func(*testing.T) *sdk
		wantErr        error
	}{
		{
			Desc:     "create test",
			giveKey:  "2",
			giveType: reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: types.BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Create(gomock.Any(),
					"/management/test_struct/2",
					`{"id":"2","create_time":22,"update_time":22,"foo":"f2","bar":"b2"}`,
				).Return(int64(1), nil)
				d.EXPECT().Get(gomock.Any(), "/management/test_struct/2").Return("", db.ErrNotFound)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
		},
		{
			Desc:     "create test (object already exist)",
			giveKey:  "2",
			giveType: reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: types.BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			wantErr: db.ErrAlreadyExist,
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/test_struct/2").Return("", nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
		},
		{
			Desc:       "prefix test",
			giveKey:    "2",
			giveType:   reflect.TypeOf(TestStruct{}),
			givePrefix: "/prefix1",
			giveValue: &TestStruct{
				BaseInfo: types.BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Create(gomock.Any(),
					"/management/prefix1/2",
					`{"id":"2","create_time":22,"update_time":22,"foo":"f2","bar":"b2"}`).
					Return(int64(1), nil)
				d.EXPECT().Get(gomock.Any(), "/management/prefix1/2").Return("", db.ErrNotFound)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			s := tc.mockNewService(t)
			q := s.New().Type(tc.giveType)
			if tc.givePrefix != "" {
				q = q.Prefix(tc.givePrefix)
			}
			obj, err := q.Create(context.TODO(), tc.giveKey, tc.giveValue)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error(), "checking error of Create")
			} else {
				assert.Nil(t, err, "checking error of Create")
				assert.Equal(t, tc.giveValue, obj, "checking Result of Create")
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		Desc           string
		giveKey        string
		giveType       reflect.Type
		giveValue      interface{}
		givePrefix     string
		createNotExist bool
		mockNewService func(*testing.T) *sdk
		wantErr        error
	}{
		{
			Desc:     "update test",
			giveKey:  "2",
			giveType: reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: types.BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/test_struct/2").
					Return(`{"id":"2","create_time":22,"update_time":11,"foo":"f2","bar":"b2"}`, nil)
				d.EXPECT().Update(gomock.Any(),
					"/management/test_struct/2",
					`{"id":"2","create_time":22,"update_time":22,"foo":"f2","bar":"b2"}`,
				).Return(int64(1), nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
		},
		{
			Desc:     "not exits and create test",
			giveKey:  "2",
			giveType: reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: types.BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			createNotExist: true,
			givePrefix:     "/prefix1",
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/prefix1/2").
					Return("", db.ErrNotFound).AnyTimes()
				d.EXPECT().Create(gomock.Any(),
					"/management/prefix1/2",
					`{"id":"2","create_time":22,"update_time":22,"foo":"f2","bar":"b2"}`,
				).Return(int64(1), nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
		},
		{
			Desc:     "not exits and failed test",
			giveKey:  "2",
			giveType: reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: types.BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			createNotExist: false,
			givePrefix:     "/prefix1",
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/prefix1/2").
					Return("", errors.New("not found"))
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantErr: errors.New("failed to get data: not found"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			s := tc.mockNewService(t)
			q := s.New().Type(tc.giveType)
			if tc.givePrefix != "" {
				q = q.Prefix(tc.givePrefix)
			}
			obj, err := q.Update(context.TODO(), tc.giveKey, tc.giveValue, tc.createNotExist)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error(), "checking error of Update")
				assert.Nil(t, obj, "checking Result of Update")
			} else {
				assert.Nil(t, err, "checking error of Update")
				assert.Equal(t, tc.giveValue, obj, "checking Result of Update")
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		Desc           string
		giveKey        string
		giveType       reflect.Type
		givePrefix     string
		mockNewService func(*testing.T) *sdk
		wantErr        error
	}{
		{
			Desc:     "delete test",
			giveKey:  "1",
			giveType: reflect.TypeOf(TestStruct{}),
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/test_struct/1").Return("", nil)
				d.EXPECT().Delete(gomock.Any(), "/management/test_struct/1").
					Return(int64(1), nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
		},
		{
			Desc:     "delete test (object not found)",
			giveKey:  "1",
			giveType: reflect.TypeOf(TestStruct{}),
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/test_struct/1").Return("", db.ErrNotFound)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantErr: db.ErrNotFound,
		},
		{
			Desc:       "prefix test",
			giveKey:    "1",
			giveType:   reflect.TypeOf(TestStruct{}),
			givePrefix: "/prefix1",
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/prefix1/1").Return("", nil)
				d.EXPECT().Delete(gomock.Any(), "/management/prefix1/1").
					Return(int64(1), nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			s := tc.mockNewService(t)
			q := s.New().Type(tc.giveType)
			if tc.givePrefix != "" {
				q = q.Prefix(tc.givePrefix)
			}
			err := q.Delete(context.TODO(), tc.giveKey)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error(), "checking error of Delete")
			} else {
				assert.Nil(t, err, "checking error of Delete")
			}
		})
	}
}

func TestPatch(t *testing.T) {
	tests := []struct {
		Desc           string
		giveKey        string
		giveType       reflect.Type
		giveValue      interface{}
		expectValue    interface{}
		givePrefix     string
		createNotExist bool
		mockNewService func(*testing.T) *sdk
		wantErr        error
	}{
		{
			Desc:     "patch object failed (not found)",
			giveKey:  "2",
			giveType: reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: types.BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/test_struct/2").
					Return("", db.ErrNotFound)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantErr: db.ErrNotFound,
		},
		{
			Desc:     "patch object failed (update failed)",
			giveKey:  "2",
			giveType: reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: types.BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Foo: "f2",
				Bar: "b2",
			},
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/test_struct/2").
					Return(`{"foo":"f1", "bar":"b1"}`, nil).AnyTimes()
				d.EXPECT().Update(gomock.Any(), "/management/test_struct/2", `{"bar":"b2","create_time":22,"foo":"f2","id":"2","update_time":22}`).Return(int64(1), errors.New("db error"))
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			wantErr: errors.New("failed to update: db error"),
		},
		{
			Desc:     "patch object failed (update successfully)",
			giveKey:  "2",
			giveType: reflect.TypeOf(TestStruct{}),
			giveValue: &TestStruct{
				BaseInfo: types.BaseInfo{
					ID:         "2",
					CreateTime: 22,
					UpdateTime: 22,
				},
				Bar: "b2",
			},
			mockNewService: func(t *testing.T) *sdk {
				ctrl := gomock.NewController(t)
				d := db.NewMockDB(ctrl)
				d.EXPECT().Get(gomock.Any(), "/management/test_struct/2").
					Return(`{"foo":"f1", "bar":"b1"}`, nil).AnyTimes()
				d.EXPECT().Update(gomock.Any(), "/management/test_struct/2", `{"bar":"b2","create_time":22,"foo":"f1","id":"2","update_time":22}`).Return(int64(1), nil)
				return &sdk{
					db:            d,
					clusterPrefix: "/management",
				}
			},
			expectValue: &TestStruct{
				BaseInfo: types.BaseInfo{
					ID:         "2",
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
			s := tc.mockNewService(t)
			q := s.New().Type(tc.giveType)
			if tc.givePrefix != "" {
				q = q.Prefix(tc.givePrefix)
			}
			obj, err := q.Patch(context.TODO(), tc.giveKey, tc.giveValue)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error(), "checking error of Update")
				assert.Nil(t, obj, "checking Result of Update")
			} else {
				assert.Nil(t, err, "checking error of Update")
				assert.Equal(t, tc.expectValue, obj, "checking Result of Update")
			}
		})
	}
}
