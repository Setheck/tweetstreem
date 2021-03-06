// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	http "net/http"

	oauth "github.com/gomodule/oauth1/oauth"
	mock "github.com/stretchr/testify/mock"

	url "net/url"
)

// OauthFacade is an autogenerated mock type for the OauthFacade type
type OauthFacade struct {
	mock.Mock
}

// AuthorizationURL provides a mock function with given fields: temporaryCredentials, additionalParams
func (_m *OauthFacade) AuthorizationURL(temporaryCredentials *oauth.Credentials, additionalParams url.Values) string {
	ret := _m.Called(temporaryCredentials, additionalParams)

	var r0 string
	if rf, ok := ret.Get(0).(func(*oauth.Credentials, url.Values) string); ok {
		r0 = rf(temporaryCredentials, additionalParams)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Get provides a mock function with given fields: client, credentials, urlStr, form
func (_m *OauthFacade) Get(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error) {
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

// OaRequest provides a mock function with given fields: method, u, conf
func (_m *OauthFacade) OaRequest(method string, u string, conf url.Values) ([]byte, error) {
	ret := _m.Called(method, u, conf)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, string, url.Values) []byte); ok {
		r0 = rf(method, u, conf)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, url.Values) error); ok {
		r1 = rf(method, u, conf)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Post provides a mock function with given fields: client, credentials, urlStr, form
func (_m *OauthFacade) Post(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error) {
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

// RequestTemporaryCredentials provides a mock function with given fields: client, callbackURL, additionalParams
func (_m *OauthFacade) RequestTemporaryCredentials(client *http.Client, callbackURL string, additionalParams url.Values) (*oauth.Credentials, error) {
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

// RequestToken provides a mock function with given fields: client, temporaryCredentials, verifier
func (_m *OauthFacade) RequestToken(client *http.Client, temporaryCredentials *oauth.Credentials, verifier string) (*oauth.Credentials, url.Values, error) {
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

// SetSecret provides a mock function with given fields: secret
func (_m *OauthFacade) SetSecret(secret string) {
	_m.Called(secret)
}

// SetToken provides a mock function with given fields: token
func (_m *OauthFacade) SetToken(token string) {
	_m.Called(token)
}
