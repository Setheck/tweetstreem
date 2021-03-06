// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"
	http "net/http"

	mock "github.com/stretchr/testify/mock"

	oauth "github.com/gomodule/oauth1/oauth"

	url "net/url"
)

// OauthClient is an autogenerated mock type for the OauthClient type
type OauthClient struct {
	mock.Mock
}

// AuthorizationHeader provides a mock function with given fields: credentials, method, u, params
func (_m *OauthClient) AuthorizationHeader(credentials *oauth.Credentials, method string, u *url.URL, params url.Values) string {
	ret := _m.Called(credentials, method, u, params)

	var r0 string
	if rf, ok := ret.Get(0).(func(*oauth.Credentials, string, *url.URL, url.Values) string); ok {
		r0 = rf(credentials, method, u, params)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AuthorizationURL provides a mock function with given fields: temporaryCredentials, additionalParams
func (_m *OauthClient) AuthorizationURL(temporaryCredentials *oauth.Credentials, additionalParams url.Values) string {
	ret := _m.Called(temporaryCredentials, additionalParams)

	var r0 string
	if rf, ok := ret.Get(0).(func(*oauth.Credentials, url.Values) string); ok {
		r0 = rf(temporaryCredentials, additionalParams)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Delete provides a mock function with given fields: client, credentials, urlStr, form
func (_m *OauthClient) Delete(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ret := _m.Called(client, credentials, urlStr, form)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(*http.Client, *oauth.Credentials, string, url.Values) *http.Response); ok {
		r0 = rf(client, credentials, urlStr, form)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Client, *oauth.Credentials, string, url.Values) error); ok {
		r1 = rf(client, credentials, urlStr, form)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteContext provides a mock function with given fields: ctx, credentials, urlStr, form
func (_m *OauthClient) DeleteContext(ctx context.Context, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ret := _m.Called(ctx, credentials, urlStr, form)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(context.Context, *oauth.Credentials, string, url.Values) *http.Response); ok {
		r0 = rf(ctx, credentials, urlStr, form)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *oauth.Credentials, string, url.Values) error); ok {
		r1 = rf(ctx, credentials, urlStr, form)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: client, credentials, urlStr, form
func (_m *OauthClient) Get(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ret := _m.Called(client, credentials, urlStr, form)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(*http.Client, *oauth.Credentials, string, url.Values) *http.Response); ok {
		r0 = rf(client, credentials, urlStr, form)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Client, *oauth.Credentials, string, url.Values) error); ok {
		r1 = rf(client, credentials, urlStr, form)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetContext provides a mock function with given fields: ctx, credentials, urlStr, form
func (_m *OauthClient) GetContext(ctx context.Context, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ret := _m.Called(ctx, credentials, urlStr, form)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(context.Context, *oauth.Credentials, string, url.Values) *http.Response); ok {
		r0 = rf(ctx, credentials, urlStr, form)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *oauth.Credentials, string, url.Values) error); ok {
		r1 = rf(ctx, credentials, urlStr, form)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Post provides a mock function with given fields: client, credentials, urlStr, form
func (_m *OauthClient) Post(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ret := _m.Called(client, credentials, urlStr, form)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(*http.Client, *oauth.Credentials, string, url.Values) *http.Response); ok {
		r0 = rf(client, credentials, urlStr, form)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Client, *oauth.Credentials, string, url.Values) error); ok {
		r1 = rf(client, credentials, urlStr, form)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PostContext provides a mock function with given fields: ctx, credentials, urlStr, form
func (_m *OauthClient) PostContext(ctx context.Context, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ret := _m.Called(ctx, credentials, urlStr, form)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(context.Context, *oauth.Credentials, string, url.Values) *http.Response); ok {
		r0 = rf(ctx, credentials, urlStr, form)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *oauth.Credentials, string, url.Values) error); ok {
		r1 = rf(ctx, credentials, urlStr, form)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Put provides a mock function with given fields: client, credentials, urlStr, form
func (_m *OauthClient) Put(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ret := _m.Called(client, credentials, urlStr, form)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(*http.Client, *oauth.Credentials, string, url.Values) *http.Response); ok {
		r0 = rf(client, credentials, urlStr, form)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Client, *oauth.Credentials, string, url.Values) error); ok {
		r1 = rf(client, credentials, urlStr, form)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PutContext provides a mock function with given fields: ctx, credentials, urlStr, form
func (_m *OauthClient) PutContext(ctx context.Context, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ret := _m.Called(ctx, credentials, urlStr, form)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(context.Context, *oauth.Credentials, string, url.Values) *http.Response); ok {
		r0 = rf(ctx, credentials, urlStr, form)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *oauth.Credentials, string, url.Values) error); ok {
		r1 = rf(ctx, credentials, urlStr, form)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RenewRequestCredentials provides a mock function with given fields: client, credentials, sessionHandle
func (_m *OauthClient) RenewRequestCredentials(client *http.Client, credentials *oauth.Credentials, sessionHandle string) (*oauth.Credentials, url.Values, error) {
	ret := _m.Called(client, credentials, sessionHandle)

	var r0 *oauth.Credentials
	if rf, ok := ret.Get(0).(func(*http.Client, *oauth.Credentials, string) *oauth.Credentials); ok {
		r0 = rf(client, credentials, sessionHandle)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth.Credentials)
		}
	}

	var r1 url.Values
	if rf, ok := ret.Get(1).(func(*http.Client, *oauth.Credentials, string) url.Values); ok {
		r1 = rf(client, credentials, sessionHandle)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(url.Values)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(*http.Client, *oauth.Credentials, string) error); ok {
		r2 = rf(client, credentials, sessionHandle)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// RenewRequestCredentialsContext provides a mock function with given fields: ctx, credentials, sessionHandle
func (_m *OauthClient) RenewRequestCredentialsContext(ctx context.Context, credentials *oauth.Credentials, sessionHandle string) (*oauth.Credentials, url.Values, error) {
	ret := _m.Called(ctx, credentials, sessionHandle)

	var r0 *oauth.Credentials
	if rf, ok := ret.Get(0).(func(context.Context, *oauth.Credentials, string) *oauth.Credentials); ok {
		r0 = rf(ctx, credentials, sessionHandle)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth.Credentials)
		}
	}

	var r1 url.Values
	if rf, ok := ret.Get(1).(func(context.Context, *oauth.Credentials, string) url.Values); ok {
		r1 = rf(ctx, credentials, sessionHandle)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(url.Values)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, *oauth.Credentials, string) error); ok {
		r2 = rf(ctx, credentials, sessionHandle)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// RequestTemporaryCredentials provides a mock function with given fields: client, callbackURL, additionalParams
func (_m *OauthClient) RequestTemporaryCredentials(client *http.Client, callbackURL string, additionalParams url.Values) (*oauth.Credentials, error) {
	ret := _m.Called(client, callbackURL, additionalParams)

	var r0 *oauth.Credentials
	if rf, ok := ret.Get(0).(func(*http.Client, string, url.Values) *oauth.Credentials); ok {
		r0 = rf(client, callbackURL, additionalParams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth.Credentials)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Client, string, url.Values) error); ok {
		r1 = rf(client, callbackURL, additionalParams)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RequestTemporaryCredentialsContext provides a mock function with given fields: ctx, callbackURL, additionalParams
func (_m *OauthClient) RequestTemporaryCredentialsContext(ctx context.Context, callbackURL string, additionalParams url.Values) (*oauth.Credentials, error) {
	ret := _m.Called(ctx, callbackURL, additionalParams)

	var r0 *oauth.Credentials
	if rf, ok := ret.Get(0).(func(context.Context, string, url.Values) *oauth.Credentials); ok {
		r0 = rf(ctx, callbackURL, additionalParams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth.Credentials)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, url.Values) error); ok {
		r1 = rf(ctx, callbackURL, additionalParams)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RequestToken provides a mock function with given fields: client, temporaryCredentials, verifier
func (_m *OauthClient) RequestToken(client *http.Client, temporaryCredentials *oauth.Credentials, verifier string) (*oauth.Credentials, url.Values, error) {
	ret := _m.Called(client, temporaryCredentials, verifier)

	var r0 *oauth.Credentials
	if rf, ok := ret.Get(0).(func(*http.Client, *oauth.Credentials, string) *oauth.Credentials); ok {
		r0 = rf(client, temporaryCredentials, verifier)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth.Credentials)
		}
	}

	var r1 url.Values
	if rf, ok := ret.Get(1).(func(*http.Client, *oauth.Credentials, string) url.Values); ok {
		r1 = rf(client, temporaryCredentials, verifier)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(url.Values)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(*http.Client, *oauth.Credentials, string) error); ok {
		r2 = rf(client, temporaryCredentials, verifier)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// RequestTokenContext provides a mock function with given fields: ctx, temporaryCredentials, verifier
func (_m *OauthClient) RequestTokenContext(ctx context.Context, temporaryCredentials *oauth.Credentials, verifier string) (*oauth.Credentials, url.Values, error) {
	ret := _m.Called(ctx, temporaryCredentials, verifier)

	var r0 *oauth.Credentials
	if rf, ok := ret.Get(0).(func(context.Context, *oauth.Credentials, string) *oauth.Credentials); ok {
		r0 = rf(ctx, temporaryCredentials, verifier)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth.Credentials)
		}
	}

	var r1 url.Values
	if rf, ok := ret.Get(1).(func(context.Context, *oauth.Credentials, string) url.Values); ok {
		r1 = rf(ctx, temporaryCredentials, verifier)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(url.Values)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, *oauth.Credentials, string) error); ok {
		r2 = rf(ctx, temporaryCredentials, verifier)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// RequestTokenXAuth provides a mock function with given fields: client, temporaryCredentials, user, password
func (_m *OauthClient) RequestTokenXAuth(client *http.Client, temporaryCredentials *oauth.Credentials, user string, password string) (*oauth.Credentials, url.Values, error) {
	ret := _m.Called(client, temporaryCredentials, user, password)

	var r0 *oauth.Credentials
	if rf, ok := ret.Get(0).(func(*http.Client, *oauth.Credentials, string, string) *oauth.Credentials); ok {
		r0 = rf(client, temporaryCredentials, user, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth.Credentials)
		}
	}

	var r1 url.Values
	if rf, ok := ret.Get(1).(func(*http.Client, *oauth.Credentials, string, string) url.Values); ok {
		r1 = rf(client, temporaryCredentials, user, password)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(url.Values)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(*http.Client, *oauth.Credentials, string, string) error); ok {
		r2 = rf(client, temporaryCredentials, user, password)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// RequestTokenXAuthContext provides a mock function with given fields: ctx, temporaryCredentials, user, password
func (_m *OauthClient) RequestTokenXAuthContext(ctx context.Context, temporaryCredentials *oauth.Credentials, user string, password string) (*oauth.Credentials, url.Values, error) {
	ret := _m.Called(ctx, temporaryCredentials, user, password)

	var r0 *oauth.Credentials
	if rf, ok := ret.Get(0).(func(context.Context, *oauth.Credentials, string, string) *oauth.Credentials); ok {
		r0 = rf(ctx, temporaryCredentials, user, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth.Credentials)
		}
	}

	var r1 url.Values
	if rf, ok := ret.Get(1).(func(context.Context, *oauth.Credentials, string, string) url.Values); ok {
		r1 = rf(ctx, temporaryCredentials, user, password)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(url.Values)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, *oauth.Credentials, string, string) error); ok {
		r2 = rf(ctx, temporaryCredentials, user, password)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// SetAuthorizationHeader provides a mock function with given fields: header, credentials, method, u, form
func (_m *OauthClient) SetAuthorizationHeader(header http.Header, credentials *oauth.Credentials, method string, u *url.URL, form url.Values) error {
	ret := _m.Called(header, credentials, method, u, form)

	var r0 error
	if rf, ok := ret.Get(0).(func(http.Header, *oauth.Credentials, string, *url.URL, url.Values) error); ok {
		r0 = rf(header, credentials, method, u, form)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SignForm provides a mock function with given fields: credentials, method, urlStr, form
func (_m *OauthClient) SignForm(credentials *oauth.Credentials, method string, urlStr string, form url.Values) error {
	ret := _m.Called(credentials, method, urlStr, form)

	var r0 error
	if rf, ok := ret.Get(0).(func(*oauth.Credentials, string, string, url.Values) error); ok {
		r0 = rf(credentials, method, urlStr, form)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SignParam provides a mock function with given fields: credentials, method, urlStr, params
func (_m *OauthClient) SignParam(credentials *oauth.Credentials, method string, urlStr string, params url.Values) {
	_m.Called(credentials, method, urlStr, params)
}
