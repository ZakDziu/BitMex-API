package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/golang/mock/gomock"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bitmex-api/pkg/authmiddleware"
	"bitmex-api/pkg/authmiddleware/mockauthmiddleware"
	"bitmex-api/pkg/model"
	"bitmex-api/pkg/model/ui/auth"
	"bitmex-api/pkg/store"
	"bitmex-api/pkg/store/mockpostgresstore"
)

var testMapAuthHandler = map[string][]model.TestStructure{
	"Login": {
		{
			Name:   "Positive",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/login",
			Data: model.AuthUser{
				Username: "user",
				Password: "password",
			},
			ExpectedData: &authmiddleware.Tokens{
				Access:  "access_token",
				Refresh: "refresh_token",
			},
			PositiveTest: true,
			WhatError:    nil,
			Mock:         makeList(AuthRepoGetByUsernameMock, MiddlewareCreateTokensMock),
			MockData: [][]interface{}{
				{
					&model.AuthUser{
						Username: "user",
						Password: authmiddleware.CreateHashPassword("password"),
					},
				},
				{
					&authmiddleware.Tokens{
						Access:  "access_token",
						Refresh: "refresh_token",
					},
				},
			},
		},
		{
			Name:         "NegativeJsonData",
			Method:       http.MethodPost,
			URL:          "https://localhost:8000/api/v1/login",
			Data:         "{",
			PositiveTest: false, WhatError: model.ErrInvalidBody,
		},
		{
			Name:   "NegativeUsernameAndPasswordEmpty",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/login",
			Data: model.AuthUser{
				Username: "",
				Password: "",
			},
			PositiveTest: false, WhatError: model.ErrInvalidBody,
		},
		{
			Name:   "NegativeAuthRepoGetByUsernameMockNotFound",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/login",
			Data: model.AuthUser{
				Username: "username",
				Password: "password",
			},
			PositiveTest: false, WhatError: model.ErrUnauthorized,
			Mock: makeList(AuthRepoGetByUsernameMock),
			MockData: [][]interface{}{
				{
					errors.New(model.NotFound),
				},
			},
		},
		{
			Name:   "NegativeAuthRepoGetByUsernameMock",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/login",
			Data: model.AuthUser{
				Username: "username",
				Password: "password",
			},
			PositiveTest: false, WhatError: model.ErrUnhealthy,
			Mock: makeList(AuthRepoGetByUsernameMock),
			MockData: [][]interface{}{
				{
					model.ErrUnhealthy,
				},
			},
		},
		{
			Name:   "NegativeIncorrectPassword",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/login",
			Data: model.AuthUser{
				Username: "username",
				Password: "password",
			},
			PositiveTest: false, WhatError: model.ErrUnauthorized,
			Mock: makeList(AuthRepoGetByUsernameMock),
			MockData: [][]interface{}{
				{
					&model.AuthUser{
						Username: "user",
						Password: authmiddleware.CreateHashPassword("incorrect-password"),
					},
				},
			},
		},
		{
			Name:   "NegativeMiddlewareCreateTokensMock",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/login",
			Data: model.AuthUser{
				Username: "username",
				Password: "password",
			},
			PositiveTest: false, WhatError: model.ErrUnhealthy,
			Mock: makeList(AuthRepoGetByUsernameMock, MiddlewareCreateTokensMock),
			MockData: [][]interface{}{
				{
					&model.AuthUser{
						Username: "user",
						Password: authmiddleware.CreateHashPassword("password"),
					},
				},
				{
					model.ErrUnhealthy,
				},
			},
		},
	},
	"Register": {
		{
			Name:   "Positive",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/registration",
			Data: model.AuthUser{
				Username: "user",
				Password: "password",
				Role:     model.AdminUserRole,
			},
			ExpectedData: auth.RegistrationResponse{
				Status: "user created",
			},
			PositiveTest: true,
			WhatError:    nil,
			Mock:         makeList(MiddlewareGetUserRoleMock, AuthRepoGetByUsernameMock, AuthRepoCreateMock),
			MockData: [][]interface{}{
				{
					model.AdminUserRole,
				},
				{
					&model.AuthUser{
						ID: uuid.Nil,
					},
				},
				{},
			},
		},
		{
			Name:         "NegativeJsonData",
			Method:       http.MethodPost,
			URL:          "https://localhost:8000/api/v1/registration",
			Data:         "{",
			PositiveTest: false, WhatError: model.ErrInvalidBody,
		},
		{
			Name:   "NegativeGetUserRole",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/registration",
			Data: model.AuthUser{
				Username: "",
				Password: "",
			},
			PositiveTest: false, WhatError: model.ErrUnauthorized,
			Mock: makeList(MiddlewareGetUserRoleMock),
			MockData: [][]interface{}{
				{
					model.ErrUnhealthy,
				},
			},
		},
		{
			Name:   "NegativeUserRole",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/registration",
			Data: model.AuthUser{
				Username: "",
				Password: "",
			},
			PositiveTest: false, WhatError: model.ErrInvalidRole,
			Mock: makeList(MiddlewareGetUserRoleMock),
			MockData: [][]interface{}{
				{
					model.BaseUserRole,
				},
			},
		},
		{
			Name:   "NegativeUsernameAndPasswordAndRoleEmpty",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/registration",
			Data: model.AuthUser{
				Username: "",
				Password: "",
				Role:     "",
			},
			PositiveTest: false, WhatError: model.ErrInvalidBody,
			Mock: makeList(MiddlewareGetUserRoleMock),
			MockData: [][]interface{}{
				{
					model.AdminUserRole,
				},
			},
		},
		{
			Name:   "NegativeAuthRepoGetByUsernameMock",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/registration",
			Data: model.AuthUser{
				Username: "username",
				Password: "password",
				Role:     model.AdminUserRole,
			},
			PositiveTest: false, WhatError: model.ErrUnhealthy,
			Mock: makeList(MiddlewareGetUserRoleMock, AuthRepoGetByUsernameMock),
			MockData: [][]interface{}{
				{
					model.AdminUserRole,
				},
				{
					model.ErrUnhealthy,
				},
			},
		},
		{
			Name:   "NegativeAuthRepoGetByUsernameMockUserExists",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/registration",
			Data: model.AuthUser{
				Username: "username",
				Password: "password",
				Role:     model.AdminUserRole,
			},
			PositiveTest: false, WhatError: model.ErrUsernameExist,
			Mock: makeList(MiddlewareGetUserRoleMock, AuthRepoGetByUsernameMock),
			MockData: [][]interface{}{
				{
					model.AdminUserRole,
				},
				{
					&model.AuthUser{ID: uuid.NewV4()},
				},
			},
		},
		{
			Name:   "NegativeAuthRepoCreateMock",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/registration",
			Data: model.AuthUser{
				Username: "username",
				Password: "password",
				Role:     model.AdminUserRole,
			},
			PositiveTest: false, WhatError: model.ErrUnhealthy,
			Mock: makeList(MiddlewareGetUserRoleMock, AuthRepoGetByUsernameMock, AuthRepoCreateMock),
			MockData: [][]interface{}{
				{
					model.AdminUserRole,
				},
				{
					&model.AuthUser{
						ID: uuid.Nil,
					},
				},
				{
					model.ErrUnhealthy,
				},
			},
		},
	},
	"Refresh": {
		{
			Name:   "Positive",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/refresh",
			Data: authmiddleware.Tokens{
				Access:  "old-token",
				Refresh: "old-token",
			},
			ExpectedData: &authmiddleware.Tokens{
				Access:  "new-token",
				Refresh: "new-token",
			},
			PositiveTest: true,
			WhatError:    nil,
			Mock:         makeList(MiddlewareRefreshTokensMock),
			MockData: [][]interface{}{
				{
					&authmiddleware.Tokens{
						Access:  "new-token",
						Refresh: "new-token",
					},
				},
			},
		},
		{
			Name:         "NegativeJsonData",
			Method:       http.MethodPost,
			URL:          "https://localhost:8000/api/v1/refresh",
			Data:         "{",
			PositiveTest: false, WhatError: model.ErrInvalidBody,
		},
		{
			Name:   "NeativeMiddlewareRefreshTokensMock",
			Method: http.MethodPost,
			URL:    "https://localhost:8000/api/v1/refresh",
			Data: authmiddleware.Tokens{
				Access:  "old-token",
				Refresh: "old-token",
			},
			PositiveTest: false, WhatError: model.ErrUnauthorized,
			Mock: makeList(MiddlewareRefreshTokensMock),
			MockData: [][]interface{}{
				{
					model.ErrUnhealthy,
				},
			},
		},
	},
	"ChangePassword": {
		{
			Name:   "Positive",
			Method: http.MethodPatch,
			URL:    "https://localhost:8000/api/v1/change-password",
			Data: model.ChangePassword{
				OldPassword: "old-pass",
				NewPassword: "new-pass",
			},
			ExpectedData: &authmiddleware.Tokens{
				Access:  "acc-token",
				Refresh: "refresh-token",
			},
			PositiveTest: true,
			WhatError:    nil,
			Mock:         makeList(MiddlewareGetUserIDMock, AuthRepoGetMock, AuthRepoChangePasswordMock, MiddlewareCreateTokensMock),
			MockData: [][]interface{}{
				{
					uuid.NewV4(),
				},
				{
					&model.AuthUser{
						Password: authmiddleware.CreateHashPassword("old-pass"),
					},
					true,
				},
				{},
				{
					&authmiddleware.Tokens{
						Access:  "acc-token",
						Refresh: "refresh-token",
					},
				},
			},
		},
		{
			Name:         "NegativeJsonData",
			Method:       http.MethodPatch,
			URL:          "https://localhost:8000/api/v1/change-password",
			Data:         "{",
			PositiveTest: false, WhatError: model.ErrInvalidBody,
		},
		{
			Name:   "NegativeOldAndNewPasswordEmpty",
			Method: http.MethodPatch,
			URL:    "https://localhost:8000/api/v1/change-password",
			Data: model.ChangePassword{
				OldPassword: "",
				NewPassword: "",
			},
			PositiveTest: false, WhatError: model.ErrInvalidBody,
		},
		{
			Name:   "NegativeMiddlewareGetUserIDMock",
			Method: http.MethodPatch,
			URL:    "https://localhost:8000/api/v1/change-password",
			Data: model.ChangePassword{
				OldPassword: "old-pass",
				NewPassword: "new-pass",
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
			Name:   "NegativeUserNotExist",
			Method: http.MethodPatch,
			URL:    "https://localhost:8000/api/v1/change-password",
			Data: model.ChangePassword{
				OldPassword: "old-pass",
				NewPassword: "new-pass",
			},
			PositiveTest: false, WhatError: model.ErrRefreshExpired,
			Mock: makeList(MiddlewareGetUserIDMock, AuthRepoGetMock),
			MockData: [][]interface{}{
				{
					uuid.NewV4(),
				},
				{
					false,
				},
			},
		},
		{
			Name:   "NegativeAuthRepoGetMockIncorrectPassword",
			Method: http.MethodPatch,
			URL:    "https://localhost:8000/api/v1/change-password",
			Data: model.ChangePassword{
				OldPassword: "old-pass",
				NewPassword: "new-pass",
			},
			PositiveTest: false, WhatError: model.ErrUnauthorized,
			Mock: makeList(MiddlewareGetUserIDMock, AuthRepoGetMock),
			MockData: [][]interface{}{
				{
					uuid.NewV4(),
				},
				{
					&model.AuthUser{
						Password: authmiddleware.CreateHashPassword("incorrect-pass"),
					},
					true,
				},
			},
		},
		{
			Name:   "NegativeAuthRepoChangePasswordMock",
			Method: http.MethodPatch,
			URL:    "https://localhost:8000/api/v1/change-password",
			Data: model.ChangePassword{
				OldPassword: "old-pass",
				NewPassword: "new-pass",
			},
			PositiveTest: false, WhatError: model.ErrUnhealthy,
			Mock: makeList(MiddlewareGetUserIDMock, AuthRepoGetMock, AuthRepoChangePasswordMock),
			MockData: [][]interface{}{
				{
					uuid.NewV4(),
				},
				{
					&model.AuthUser{
						Password: authmiddleware.CreateHashPassword("old-pass"),
					},
					true,
				},
				{model.ErrUnhealthy},
			},
		},
		{
			Name:   "NegativeAuthRepoChangePasswordMock",
			Method: http.MethodPatch,
			URL:    "https://localhost:8000/api/v1/change-password",
			Data: model.ChangePassword{
				OldPassword: "old-pass",
				NewPassword: "new-pass",
			},
			PositiveTest: false, WhatError: model.ErrUnhealthy,
			Mock: makeList(MiddlewareGetUserIDMock, AuthRepoGetMock, AuthRepoChangePasswordMock, MiddlewareCreateTokensMock),
			MockData: [][]interface{}{
				{
					uuid.NewV4(),
				},
				{
					&model.AuthUser{
						Password: authmiddleware.CreateHashPassword("old-pass"),
					},
					true,
				},
				{},
				{
					model.ErrUnhealthy,
				},
			},
		},
	},
}

func TestAuthHandlers(t *testing.T) {
	var repos []interface{}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockAuthMiddleware := mockauthmiddleware.NewMockAuthMiddleware(mockCtrl)
	repos = append(repos, mockAuthMiddleware)

	mockPostgresStore := &store.Store{}

	testAPI := initTestAPI(t, mockAuthMiddleware, mockPostgresStore)
	mockAuthMiddleware.EXPECT().Authorize(gomock.Any()).Return().AnyTimes()

	//all repos mock what need for tests
	userAuthRepo := mockpostgresstore.NewMockAuthRepository(mockCtrl)
	mockPostgresStore.Auth = userAuthRepo
	repos = append(repos, userAuthRepo)

	//execute tests
	for apiName, testsAuthHandlers := range testMapAuthHandler {
		t.Run(apiName, func(t *testing.T) {
			for _, data := range testsAuthHandlers {
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

func MiddlewareCreateTokensMock(repos []interface{}, data []interface{}) {
	var middlewareMock *mockauthmiddleware.MockAuthMiddleware
	var result *authmiddleware.Tokens
	var err error

	for _, r := range repos {
		switch t := r.(type) {
		case *mockauthmiddleware.MockAuthMiddleware:
			middlewareMock = t
		}
	}

	for _, i := range data {
		switch t := i.(type) {
		case error:
			err = t
		case *authmiddleware.Tokens:
			result = t
		default:
			continue
		}
	}

	middlewareMock.EXPECT().CreateTokens(gomock.Any(), gomock.Any()).Return(result, err).Times(1)
}

func MiddlewareRefreshTokensMock(repos []interface{}, data []interface{}) {
	var middlewareMock *mockauthmiddleware.MockAuthMiddleware
	var result *authmiddleware.Tokens
	var err error

	for _, r := range repos {
		switch t := r.(type) {
		case *mockauthmiddleware.MockAuthMiddleware:
			middlewareMock = t
		}
	}

	for _, i := range data {
		switch t := i.(type) {
		case error:
			err = t
		case *authmiddleware.Tokens:
			result = t
		default:
			continue
		}
	}

	middlewareMock.EXPECT().Refresh(gomock.Any()).Return(result, err).Times(1)
}

func MiddlewareGetUserRoleMock(repos []interface{}, data []interface{}) {
	var middlewareMock *mockauthmiddleware.MockAuthMiddleware
	var result model.UserRole
	var err error

	for _, r := range repos {
		switch t := r.(type) {
		case *mockauthmiddleware.MockAuthMiddleware:
			middlewareMock = t
		}
	}

	for _, i := range data {
		switch t := i.(type) {
		case error:
			err = t
		case model.UserRole:
			result = t
		default:
			continue
		}
	}

	middlewareMock.EXPECT().GetUserRole(gomock.Any()).Return(result, err).Times(1)
}

func MiddlewareGetUserIDMock(repos []interface{}, data []interface{}) {
	var middlewareMock *mockauthmiddleware.MockAuthMiddleware
	var result uuid.UUID
	var err error

	for _, r := range repos {
		switch t := r.(type) {
		case *mockauthmiddleware.MockAuthMiddleware:
			middlewareMock = t
		}
	}

	for _, i := range data {
		switch t := i.(type) {
		case error:
			err = t
		case uuid.UUID:
			result = t
		default:
			continue
		}
	}

	middlewareMock.EXPECT().GetUserID(gomock.Any()).Return(result, err).Times(1)
}

func AuthRepoGetByUsernameMock(repos []interface{}, data []interface{}) {
	var authMock *mockpostgresstore.MockAuthRepository
	var result *model.AuthUser
	var err error

	for _, r := range repos {
		switch t := r.(type) {
		case *mockpostgresstore.MockAuthRepository:
			authMock = t
		}
	}

	for _, i := range data {
		switch t := i.(type) {
		case error:
			err = t
		case *model.AuthUser:
			result = t
		default:
			continue
		}
	}

	authMock.EXPECT().GetByUsername(gomock.Any()).Return(result, err).Times(1)
}

func AuthRepoGetMock(repos []interface{}, data []interface{}) {
	var authMock *mockpostgresstore.MockAuthRepository
	var result *model.AuthUser
	var exist bool

	for _, r := range repos {
		switch t := r.(type) {
		case *mockpostgresstore.MockAuthRepository:
			authMock = t
		}
	}

	for _, i := range data {
		switch t := i.(type) {
		case bool:
			exist = t
		case *model.AuthUser:
			result = t
		default:
			continue
		}
	}

	authMock.EXPECT().Get(gomock.Any()).Return(result, exist).Times(1)
}

func AuthRepoChangePasswordMock(repos []interface{}, data []interface{}) {
	var authMock *mockpostgresstore.MockAuthRepository
	var err error

	for _, r := range repos {
		switch t := r.(type) {
		case *mockpostgresstore.MockAuthRepository:
			authMock = t
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

	authMock.EXPECT().ChangePassword(gomock.Any(), gomock.Any()).Return(err).Times(1)
}

func AuthRepoCreateMock(repos []interface{}, data []interface{}) {
	var authMock *mockpostgresstore.MockAuthRepository
	var err error

	for _, r := range repos {
		switch t := r.(type) {
		case *mockpostgresstore.MockAuthRepository:
			authMock = t
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

	authMock.EXPECT().Create(gomock.Any()).Return(err).Times(1)
}
