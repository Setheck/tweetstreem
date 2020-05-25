package auth

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/Setheck/tweetstreem/auth/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_DefaultOaFacade(t *testing.T) {
	oaconf := OauthConfig{
		AppToken:  "anAppToken",
		AppSecret: "anAppSecret",
		Token:     "testToken",
		Secret:    "testSecret",
	}
	dfof := NewDefaultOaFacade(oaconf)

	assert.Equal(t, oaconf.Token, dfof.Token)
	assert.Equal(t, oaconf.Secret, dfof.Secret)

	newToken := "nextToken"
	newSecret := "nextSecret"
	dfof.SetToken(newToken)
	dfof.SetSecret(newSecret)

	assert.Equal(t, newToken, dfof.Token)
	assert.Equal(t, newSecret, dfof.Secret)
}

func TestDefaultOaFacade_OaRequest(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		statusCode   int
		requestError bool
	}{
		{"get success", http.MethodGet, http.StatusOK, false},
		{"get requestError", http.MethodGet, http.StatusOK, true},
		{"get 404", http.MethodGet, http.StatusNotFound, false},
		{"post success", http.MethodPost, http.StatusOK, false},
		{"post requestError", http.MethodPost, http.StatusOK, true},
		{"post 404", http.MethodPost, http.StatusNotFound, false},
		{"put", http.MethodPut, http.StatusOK, true},
	}
	for _, test := range tests {
		theUrl := "https://example.com/asdf123"
		theBody := "this is a test body"

		var nilHttpClient *http.Client
		t.Run(test.name, func(t *testing.T) {
			oaconf := url.Values{}
			body := ioutil.NopCloser(bytes.NewBuffer([]byte(theBody)))
			resp := &http.Response{StatusCode: test.statusCode, Body: body}
			var requestError error
			if test.requestError {
				requestError = assert.AnError
			}
			mockOauthClient := new(mocks.OauthClient)
			switch test.method {
			case http.MethodGet:
				mockOauthClient.On("Get", nilHttpClient, mock.AnythingOfType("*oauth.Credentials"), theUrl, mock.AnythingOfType("url.Values")).
					Return(resp, requestError)
			case http.MethodPost:
				mockOauthClient.On("Post", nilHttpClient, mock.AnythingOfType("*oauth.Credentials"), theUrl, mock.AnythingOfType("url.Values")).
					Return(resp, requestError)
			}

			dfac := NewDefaultOaFacade(OauthConfig{})
			dfac.OauthClient = mockOauthClient

			output, err := dfac.OaRequest(test.method, theUrl, oaconf)
			if !test.requestError && test.statusCode == http.StatusOK {
				assert.Equal(t, []byte(theBody), output)
			} else {
				assert.Error(t, err)
			}

			mockOauthClient.AssertExpectations(t)
		})
	}
}
