package client

import (
	"net/http"
	"testing"

	"github.com/cecobask/imdb-trakt-sync/pkg/entities"
	"github.com/cecobask/imdb-trakt-sync/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestImdbClient_doRequest(t *testing.T) {
	type args struct {
		requestFields requestFields
	}
	dummyRequestFields := requestFields{
		Method:   http.MethodGet,
		Endpoint: "/",
	}
	tests := []struct {
		name         string
		args         args
		expectations func(*testing.T) (string, func())
		assertions   func(*testing.T, *http.Response, error)
	}{
		{
			name: "handle status ok",
			args: args{
				requestFields: dummyRequestFields,
			},
			expectations: func(t *testing.T) (string, func()) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					if r.Method != dummyRequestFields.Method || r.URL.Path != dummyRequestFields.Endpoint {
						t.Error("http request does not match expectations")
					}
					w.WriteHeader(http.StatusOK)
				}
				return testutils.NewHttpTestServer(handler)
			},
			assertions: func(t *testing.T, res *http.Response, err error) {
				assert.NotNil(t, res)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, res.StatusCode)
			},
		},
		{
			name: "handle status not found",
			args: args{
				requestFields: dummyRequestFields,
			},
			expectations: func(t *testing.T) (string, func()) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					if r.Method != dummyRequestFields.Method || r.URL.Path != dummyRequestFields.Endpoint {
						t.Error("http request does not match expectations")
					}
					w.WriteHeader(http.StatusNotFound)
				}
				return testutils.NewHttpTestServer(handler)
			},
			assertions: func(t *testing.T, res *http.Response, err error) {
				assert.NotNil(t, res)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusNotFound, res.StatusCode)
			},
		},
		{
			name: "handle status forbidden",
			args: args{
				requestFields: dummyRequestFields,
			},
			expectations: func(t *testing.T) (string, func()) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					if r.Method != dummyRequestFields.Method || r.URL.Path != dummyRequestFields.Endpoint {
						t.Error("http request does not match expectations")
					}
					w.WriteHeader(http.StatusForbidden)
				}
				return testutils.NewHttpTestServer(handler)
			},
			assertions: func(t *testing.T, res *http.Response, err error) {
				assert.Nil(t, res)
				assert.Error(t, err)
			},
		},
		{
			name: "handle unexpected status",
			args: args{
				requestFields: dummyRequestFields,
			},
			expectations: func(t *testing.T) (string, func()) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					if r.Method != dummyRequestFields.Method || r.URL.Path != dummyRequestFields.Endpoint {
						t.Error("http request does not match expectations")
					}
					w.WriteHeader(http.StatusInternalServerError)
				}
				return testutils.NewHttpTestServer(handler)
			},
			assertions: func(t *testing.T, res *http.Response, err error) {
				assert.Nil(t, res)
				assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverUrl, cleanup := tt.expectations(t)
			defer cleanup()
			tt.args.requestFields.BasePath = serverUrl
			c := &ImdbClient{
				client: http.DefaultClient,
			}
			res, err := c.doRequest(tt.args.requestFields)
			tt.assertions(t, res, err)
		})
	}
}

func TestImdbClient_ListGet(t *testing.T) {
	type args struct {
		listId string
	}
	tests := []struct {
		name         string
		args         args
		expectations func(*testing.T) (string, func())
		assertions   func(*testing.T, *entities.ImdbList, error)
	}{
		{
			name: "successfully get list",
			args: args{
				listId: "ls123456",
			},
			expectations: func(t *testing.T) (string, func()) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet || r.URL.Path != "/list/ls123456/export" {
						t.Error("http request does not match expectations")
					}
					w.Header().Set(imdbHeaderKeyContentDisposition, `attachment; filename="Watched (2023).csv"`)
					w.WriteHeader(http.StatusOK)
					if err := testutils.PopulateHttpResponseWithFileContents(w, "testdata/imdb_list.csv"); err != nil {
						t.Error(err)
					}
				}
				return testutils.NewHttpTestServer(handler)
			},
			assertions: func(t *testing.T, list *entities.ImdbList, err error) {
				assert.NotNil(t, list)
				assert.NoError(t, err)
				assert.Equal(t, "ls123456", list.ListId)
				assert.Equal(t, "Watched (2023)", list.ListName)
				assert.Equal(t, 3, len(list.ListItems))
				assert.Equal(t, false, list.IsWatchlist)
				assert.Equal(t, "watched-2023", list.TraktListSlug)
			},
		},
		{
			name: "handle error when list is not found",
			args: args{
				listId: "ls123456",
			},
			expectations: func(t *testing.T) (string, func()) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet || r.URL.Path != "/list/ls123456/export" {
						t.Error("http request does not match expectations")
					}
					w.WriteHeader(http.StatusNotFound)
				}
				return testutils.NewHttpTestServer(handler)
			},
			assertions: func(t *testing.T, list *entities.ImdbList, err error) {
				assert.Nil(t, list)
				assert.Error(t, err)
				assert.IsType(t, new(ApiError), err)
				assert.ErrorContains(t, err, "could not be found")
			},
		},
		{
			name: "handle unexpected status",
			args: args{
				listId: "ls123456",
			},
			expectations: func(t *testing.T) (string, func()) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet || r.URL.Path != "/list/ls123456/export" {
						t.Error("http request does not match expectations")
					}
					w.WriteHeader(http.StatusInternalServerError)
				}
				return testutils.NewHttpTestServer(handler)
			},
			assertions: func(t *testing.T, list *entities.ImdbList, err error) {
				assert.Nil(t, list)
				assert.Error(t, err)
				assert.IsType(t, new(ApiError), err)
				assert.ErrorContains(t, err, "unexpected status code")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverUrl, cleanup := tt.expectations(t)
			defer cleanup()
			c := &ImdbClient{
				client: http.DefaultClient,
				config: ImdbConfig{
					BasePath: serverUrl,
				},
			}
			list, err := c.ListGet(tt.args.listId)
			tt.assertions(t, list, err)
		})
	}
}

func TestImdbClient_WatchlistGet(t *testing.T) {
	tests := []struct {
		name         string
		expectations func(*testing.T) (string, func())
		assertions   func(*testing.T, *entities.ImdbList, error)
	}{
		{
			name: "successfully get watchlist",
			expectations: func(t *testing.T) (string, func()) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet || r.URL.Path != "/list/ls123456/export" {
						t.Error("http request does not match expectations")
					}
					w.Header().Set(imdbHeaderKeyContentDisposition, `attachment; filename="WATCHLIST.csv"`)
					w.WriteHeader(http.StatusOK)
					if err := testutils.PopulateHttpResponseWithFileContents(w, "testdata/imdb_list.csv"); err != nil {
						t.Error(err)
					}
				}
				return testutils.NewHttpTestServer(handler)
			},
			assertions: func(t *testing.T, list *entities.ImdbList, err error) {
				assert.NotNil(t, list)
				assert.NoError(t, err)
				assert.Equal(t, "ls123456", list.ListId)
				assert.Equal(t, "WATCHLIST", list.ListName)
				assert.Equal(t, 3, len(list.ListItems))
				assert.Equal(t, true, list.IsWatchlist)
				assert.Equal(t, "watchlist", list.TraktListSlug)
			},
		},
		{
			name: "fail to get watchlist",
			expectations: func(t *testing.T) (string, func()) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet || r.URL.Path != "/list/ls123456/export" {
						t.Error("http request does not match expectations")
					}
					w.WriteHeader(http.StatusNotFound)
				}
				return testutils.NewHttpTestServer(handler)
			},
			assertions: func(t *testing.T, list *entities.ImdbList, err error) {
				assert.Nil(t, list)
				assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverUrl, cleanup := tt.expectations(t)
			defer cleanup()
			c := &ImdbClient{
				client: http.DefaultClient,
				config: ImdbConfig{
					BasePath:    serverUrl,
					WatchlistId: "ls123456",
				},
			}
			list, err := c.WatchlistGet()
			tt.assertions(t, list, err)
		})
	}
}
