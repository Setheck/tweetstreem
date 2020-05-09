package app

import (
	"net/http"
	"net/url"

	"github.com/gomodule/oauth1/oauth"
)

type OauthFacade interface {
	RequestTemporaryCredentials(client *http.Client, callbackURL string, additionalParams url.Values) (*oauth.Credentials, error)
	AuthorizationURL(temporaryCredentials *oauth.Credentials, additionalParams url.Values) string
	RequestToken(client *http.Client, temporaryCredentials *oauth.Credentials, verifier string) (*oauth.Credentials, url.Values, error)
	Get(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
	Post(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
}

var _ OauthFacade = &oauth.Client{}

type OauthConfig struct {
	TemporaryCredentialRequestURI string
	TokenRequestURI               string
	ResourceOwnerAuthorizationURI string
	Credentials                   oauth.Credentials
}

func NewOaClient(c OauthConfig) *oauth.Client {
	return &oauth.Client{
		TemporaryCredentialRequestURI: c.TemporaryCredentialRequestURI,
		TokenRequestURI:               c.TokenRequestURI,
		ResourceOwnerAuthorizationURI: c.ResourceOwnerAuthorizationURI,
		Credentials:                   c.Credentials,
	}
}
