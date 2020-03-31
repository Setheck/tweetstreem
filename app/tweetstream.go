package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/gomodule/oauth1/oauth"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./.tweetstream/")

	AppToken = os.Getenv("APP_TOKEN")
	AppSecret = os.Getenv("APP_SECRET")
}

var (
	TwitterCredentialRequestURI = "https://api.twitter.com/oauth/request_token"
	TwitterTokenRequestURI      = "https://api.twitter.com/oauth/access_token"
	TwitterAuthorizeURI         = "https://api.twitter.com/oauth/authorize"

	AppToken  = ""
	AppSecret = ""
)

type TweetStream struct {
	TokenCredentials *oauth.Credentials

	oauthClient oauth.Client
}

func NewTweetStream() *TweetStream {
	return &TweetStream{
		oauthClient: oauth.Client{
			TemporaryCredentialRequestURI: TwitterCredentialRequestURI,
			TokenRequestURI:               TwitterTokenRequestURI,
			ResourceOwnerAuthorizationURI: TwitterAuthorizeURI,
			Credentials: oauth.Credentials{
				Token:  AppToken,
				Secret: AppSecret},
		},
	}
}

func (t *TweetStream) loadConfig() {
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Failed to read config file:", err)
		return
	}

	err = viper.Unmarshal(&t)
	if err != nil {
		fmt.Println(err)
	}
}

func (t *TweetStream) SaveConfig() {
	viper.Set("TokenCredentials", t.TokenCredentials)
	if err := viper.WriteConfig(); err != nil {
		log.Println(err)
	}
}

func (t *TweetStream) Authorize() error {
	tempCred, err := t.oauthClient.RequestTemporaryCredentials(nil, "oob", nil)
	if err != nil {
		return err
	}

	u := t.oauthClient.AuthorizationURL(tempCred, nil)
	OpenBrowser(u)

	var code string
	fmt.Printf("Enter Pin:")
	fmt.Scanln(&code)

	tokenCred, _, err := t.oauthClient.RequestToken(nil, tempCred, code)
	if err != nil {
		return err
	}
	t.TokenCredentials = tokenCred
	fmt.Println("Creds:", tokenCred)
	return nil
}

func (t *TweetStream) GetHomeTimeline() {
	resp, err := t.oauthClient.Get(nil, t.TokenCredentials,
		"https://api.twitter.com/1.1/statuses/home_timeline.json", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		log.Fatal(err)
	}
}

func (t *TweetStream) Init() error {
	t.loadConfig()
	if t.TokenCredentials == nil {
		return t.Authorize()
	}
	return nil
}

func (t *TweetStream) Stop() error {
	t.SaveConfig()
	return nil
}

func OpenBrowser(url string) {
	log.Println("opening url in browser:", url)
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
