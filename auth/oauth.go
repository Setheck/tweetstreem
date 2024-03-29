package auth

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gomodule/oauth1/oauth"
)

var ErrUnsupportedMethod = fmt.Errorf("unsupported method")

type OauthFacade interface {
	RequestTemporaryCredentials(client *http.Client, callbackURL string, additionalParams url.Values) (*oauth.Credentials, error)
	AuthorizationURL(temporaryCredentials *oauth.Credentials, additionalParams url.Values) string
	RequestToken(client *http.Client, temporaryCredentials *oauth.Credentials, verifier string) (*oauth.Credentials, url.Values, error)
	OaRequest(method, u string, conf url.Values) ([]byte, error)
	SetToken(token string)
	SetSecret(secret string)
	Get(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
	Post(client *http.Client, credentials *oauth.Credentials, urlStr string, form url.Values) (*http.Response, error)
}

type OauthConfig struct {
	TemporaryCredentialRequestURI string
	TokenRequestURI               string
	ResourceOwnerAuthorizationURI string
	AppToken                      string
	AppSecret                     string
	Token                         string
	Secret                        string
	UserAgent                     string
}

var _ OauthFacade = &DefaultOaFacade{}

type DefaultOaFacade struct {
	OauthClient
	UserAgent string
	Token     string
	Secret    string
}

func NewDefaultOaFacade(c OauthConfig) *DefaultOaFacade {
	client := &oauth.Client{
		TemporaryCredentialRequestURI: c.TemporaryCredentialRequestURI,
		TokenRequestURI:               c.TokenRequestURI,
		ResourceOwnerAuthorizationURI: c.ResourceOwnerAuthorizationURI,
		Credentials:                   oauth.Credentials{Token: c.AppToken, Secret: c.AppSecret},
	}
	return &DefaultOaFacade{
		OauthClient: client,
		UserAgent:   c.UserAgent,
		Token:       c.Token,
		Secret:      c.Secret,
	}
}

func (o *DefaultOaFacade) SetToken(token string) {
	o.Token = token
}

func (o *DefaultOaFacade) SetSecret(secret string) {
	o.Secret = secret
}

func (o *DefaultOaFacade) OaRequest(method, u string, conf url.Values) ([]byte, error) {
	cred := &oauth.Credentials{Token: o.Token, Secret: o.Secret}
	var resp *http.Response
	var err error
	conf.Set("User-Agent", o.UserAgent)
	switch strings.ToUpper(method) {
	case http.MethodPost:
		resp, err = o.OauthClient.Post(nil, cred, u, conf)
	case http.MethodGet:
		resp, err = o.OauthClient.Get(nil, cred, u, conf)
	}
	if err != nil {
		return nil, err
	}
	if resp != nil {
		if resp.StatusCode != http.StatusOK { // TODO: only non 200 ?
			return nil, fmt.Errorf("failed: %d - %s", resp.StatusCode, resp.Status)
		}
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	}
	return nil, ErrUnsupportedMethod
}
