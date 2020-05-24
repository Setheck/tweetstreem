package app

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gomodule/oauth1/oauth"
)

type OauthClient interface {
	SignForm(credentials *oauth.Credentials, method, urlStr string, form url.Values) error
	SignParam(credentials *oauth.Credentials, method, urlStr string, params url.Values)
	AuthorizationHeader(credentials *oauth.Credentials, method string, u *url.URL, params url.Values) string
	SetAuthorizationHeader(header http.Header, credentials *oauth.Credentials, method string, u *url.URL, form url.Values) error
	Get(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
	GetContext(ctx context.Context, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
	Post(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
	PostContext(ctx context.Context, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
	Delete(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
	DeleteContext(ctx context.Context, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
	Put(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
	PutContext(ctx context.Context, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
	RequestTemporaryCredentials(client *http.Client, callbackURL string, additionalParams url.Values) (*oauth.Credentials, error)
	RequestTemporaryCredentialsContext(ctx context.Context, callbackURL string, additionalParams url.Values) (*oauth.Credentials, error)
	RequestToken(client *http.Client, temporaryCredentials *oauth.Credentials, verifier string) (*oauth.Credentials, url.Values, error)
	RequestTokenContext(ctx context.Context, temporaryCredentials *oauth.Credentials, verifier string) (*oauth.Credentials, url.Values, error)
	RenewRequestCredentials(client *http.Client, credentials *oauth.Credentials, sessionHandle string) (*oauth.Credentials, url.Values, error)
	RenewRequestCredentialsContext(ctx context.Context, credentials *oauth.Credentials, sessionHandle string) (*oauth.Credentials, url.Values, error)
	RequestTokenXAuth(client *http.Client, temporaryCredentials *oauth.Credentials, user, password string) (*oauth.Credentials, url.Values, error)
	RequestTokenXAuthContext(ctx context.Context, temporaryCredentials *oauth.Credentials, user, password string) (*oauth.Credentials, url.Values, error)
	AuthorizationURL(temporaryCredentials *oauth.Credentials, additionalParams url.Values) string
}
