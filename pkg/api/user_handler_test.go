package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/golang/mock/gomock"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bitmex-api/pkg/authmiddleware/mockauthmiddleware"
	"bitmex-api/pkg/model"
	"bitmex-api/pkg/store"
	"bitmex-api/pkg/store/mockpostgresstore"
)

var testMapUserHandler = map[string][]model.TestStructure{
	"UpdateInfo": {
		{
			Name:   "Positive",
			Method: http.MethodPatch,
			URL:    "https://localhost:8000/api/v1/user/update-info",
			Data: model.User{
				Name:    "Name",
				Surname: "Surname",
				Phone:   "Phone",
				Address: "Address",
			},
			ExpectedData: &model.User{
				Name:    "Name",
				Surname: "Surname",
				Phone:   "Phone",
				Address: "Address",
			},
			PositiveTest: true,
			WhatError:    nil,
			Mock:         makeList(MiddlewareGetUserIDMock, UserRepoUpdateInfoMock),
			MockData: [][]interface{}{
				{
					uuid.NewV4(),
				},
				{},
			},
		},
		{
			Name:         "NegativeJsonData",
			Method:       http.MethodPatch,
			URL:          "https://localhost:8000/api/v1/user/update-info",
			Data:         "{",
			PositiveTest: false, WhatError: model.ErrInvalidBody,
		},
		{
			Name:   "NegativeMiddlewareGetUserIDMock",
			Method: http.MethodPatch,
			URL:    "https://localhost:8000/api/v1/user/update-info",
			Data: model.User{
				Name:    "Name",
				Surname: "Surname",
				Phone:   "Phone",
				Address: "Address",
			},
			PositiveTest: false, WhatError: model.ErrUnauthorized,
			Mock: makeList(MiddlewareGetUserIDMock),
			MockData: [][]interface{}{
				{
					model.ErrUnhealthy,
				},
			},
		},
		{
			Name:   "NegativeMiddlewareGetUserIDMock",
			Method: http.MethodPatch,
			URL:    "https://localhost:8000/api/v1/user/update-info",
			Data: model.User{
				Name:    "Name",
				Surname: "Surname",
				Phone:   "Phone",
				Address: "Address",
			},
			PositiveTest: false, WhatError: model.ErrUnhealthy,
			Mock: makeList(MiddlewareGetUserIDMock, UserRepoUpdateInfoMock),
			MockData: [][]interface{}{
				{
					uuid.NewV4(),
				},
				{
					model.ErrUnhealthy,
				},
			},
		},
	},
	"Get": {
		{
			Name:   "Positive",
			Method: http.MethodGet,
			URL:    "https://localhost:8000/api/v1/user/",
			ExpectedData: &model.User{
				Name:    "Name",
				Surname: "Surname",
				Phone:   "Phone",
				Address: "Address",
			},
			PositiveTest: true,
			WhatError:    nil,
			Mock:         makeList(MiddlewareGetUserIDMock, UserRepoGetMock),
			MockData: [][]interface{}{
				{
					uuid.NewV4(),
				},
				{
					&model.User{
						Name:    "Name",
						Surname: "Surname",
						Phone:   "Phone",
						Address: "Address",
					},
				},
			},
		},
		{
			Name:         "NegativeMiddlewareGetUserIDMock",
			Method:       http.MethodGet,
			URL:          "https://localhost:8000/api/v1/user/",
			PositiveTest: false, WhatError: model.ErrUnauthorized,
			Mock: makeList(MiddlewareGetUserIDMock),
			MockData: [][]interface{}{
				{
					model.ErrUnhealthy,
				},
			},
		},
		{
			Name:         "NegativeUserRepoGetMock",
			Method:       http.MethodGet,
			URL:          "https://localhost:8000/api/v1/user/",
			PositiveTest: false, WhatError: model.ErrUnhealthy,
			Mock: makeList(MiddlewareGetUserIDMock, UserRepoGetMock),
			MockData: [][]interface{}{
				{
					uuid.NewV4(),
				},
				{
					model.ErrUnhealthy,
				},
			},
		},
	},
}

func TestUserHandlers(t *testing.T) {
	var repos []interface{}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockAuthMiddleware := mockauthmiddleware.NewMockAuthMiddleware(mockCtrl)
	repos = append(repos, mockAuthMiddleware)

	mockPostgresStore := &store.Store{}

	testAPI := initTestAPI(t, mockAuthMiddleware, mockPostgresStore)
	mockAuthMiddleware.EXPECT().Authorize(gomock.Any()).Return().AnyTimes()

	//all repos mock what need for tests
	promoUserRepo := mockpostgresstore.NewMockUserRepository(mockCtrl)
	mockPostgresStore.User = promoUserRepo
	repos = append(repos, promoUserRepo)

	//execute tests
	for apiName, testsUserHandlers := range testMapUserHandler {
		t.Run(apiName, func(t *testing.T) {
			for _, data := range testsUserHandlers {
				t.Run(data.Name, func(t *testing.T) {
					if data.Mock != nil {
						for i, m := range data.Mock {
							m(repos, data.MockData[i])
						}
					}
					body, err := json.Marshal(data.Data)
					require.NoError(t, err)
					req, err := http.NewRequest(data.Method, data.URL, bytes.NewBuffer(body))
					require.NoError(t, err)

					rr := httptest.NewRecorder()
					testAPI.ServeHTTP(rr, req)

					if data.PositiveTest {
						assert.Equal(t, http.StatusOK, rr.Code, "handler return wrong status code")
						if reflect.TypeOf(data.ExpectedData).String() != "string" { // if func don't have answered or answer is string
							body, err = json.Marshal(data.ExpectedData)
							require.NoError(t, err)
							if data.SkipFields != nil {
								var bodyActual, bodyExpectet interface{}
								if data.SkipRoot != "" {
									expect := data.ExpectedData.(map[string]interface{})
									body, err = json.Marshal(expect[data.SkipRoot])
									require.NoError(t, err)
									var actualMap, expectMap []map[string]interface{}
									err = json.Unmarshal(body, &expectMap)
									require.NoError(t, err)
									keys := make([]string, 0, len(expect))
									for k := range expect {
										keys = append(keys, k)
									}
									sort.Strings(keys)
									for _, key := range keys {
										if key != data.SkipRoot {
											expectMap = append(expectMap, map[string]interface{}{key: expect[key]})
										}
									}

									var actual map[string]interface{}
									err = json.Unmarshal(rr.Body.Bytes(), &actual)
									require.NoError(t, err)
									body, err = json.Marshal(actual[data.SkipRoot])
									require.NoError(t, err)
									err = json.Unmarshal(body, &actualMap)
									require.NoError(t, err)
									keys = make([]string, 0, len(actual))
									for k := range actual {
										keys = append(keys, k)
									}
									sort.Strings(keys)
									for _, key := range keys {
										if key != data.SkipRoot {
											actualMap = append(actualMap, map[string]interface{}{key: actual[key]})
										}
									}

									for _, skip := range data.SkipFields {
										require.Equal(t, len(actualMap), len(expectMap))
										for i := range actualMap {
											delete(actualMap[i], skip)
											delete(expectMap[i], skip)
										}
									}
									bodyActual = actualMap
									bodyExpectet = expectMap
								} else {
									var actual, expect map[string]interface{}
									err = json.Unmarshal(body, &expect)
									require.NoError(t, err)
									err = json.Unmarshal(rr.Body.Bytes(), &actual)
									require.NoError(t, err)
									for _, skip := range data.SkipFields {
										delete(actual, skip)
										delete(expect, skip)
									}
									bodyActual = actual
									bodyExpectet = expect
								}
								body, err = json.Marshal(bodyExpectet)
								require.NoError(t, err)
								resActual, err := json.Marshal(bodyActual)
								require.NoError(t, err)
								rr.Body = bytes.NewBuffer(resActual)
							}
							assert.JSONEq(t, string(body), rr.Body.String())
						} else {
							assert.Equal(t, data.ExpectedData, rr.Body.String())
						}
					} else {
						body, err = json.Marshal(data.WhatError)
						require.NoError(t, err)
						assert.JSONEq(t, string(body), rr.Body.String())
					}
				})
			}
		})
	}
}

func UserRepoUpdateInfoMock(repos []interface{}, data []interface{}) {
	var userRepoMock *mockpostgresstore.MockUserRepository
	var err error

	for _, r := range repos {
		switch t := r.(type) {
		case *mockpostgresstore.MockUserRepository:
			userRepoMock = t
		}
	}

	for _, i := range data {
		switch t := i.(type) {
		case error:
			err = t
		default:
			continue
		}
	}

	userRepoMock.EXPECT().Update(gomock.Any()).Return(err).Times(1)
}

func UserRepoGetMock(repos []interface{}, data []interface{}) {
	var userRepoMock *mockpostgresstore.MockUserRepository
	var result *model.User
	var err error

	for _, r := range repos {
		switch t := r.(type) {
		case *mockpostgresstore.MockUserRepository:
			userRepoMock = t
		}
	}

	for _, i := range data {
		switch t := i.(type) {
		case error:
			err = t
		case *model.User:
			result = t
		default:
			continue
		}
	}

	userRepoMock.EXPECT().Get(gomock.Any()).Return(result, err).Times(1)
}
