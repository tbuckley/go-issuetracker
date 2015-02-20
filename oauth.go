package main

import (
	"encoding/json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
)

type WebSecrets struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}
type ClientSecrets struct {
	Web WebSecrets `json:"web"`
}

func exchangeCode(config *oauth2.Config, code string) (*oauth2.Token, *http.Client, error) {
	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, nil, err
	}
	client := config.Client(oauth2.NoContext, token)
	return token, client, nil
}

func GetOauthURL(config *oauth2.Config) string {
	return config.AuthCodeURL("foobar", oauth2.AccessTypeOffline)
}

type Authorization struct {
	Client *http.Client
	Token  *oauth2.Token
}

func StartOauthHandler(config *oauth2.Config) chan *Authorization {
	authorizationChan := make(chan *Authorization)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := GetOauthURL(config)
		w.Write([]byte("<a href=\"" + url + "\">Click here</a>"))
	})
	http.HandleFunc("/oauth", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, client, err := exchangeCode(config, code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		authorizationChan <- &Authorization{
			Client: client,
			Token:  token,
		}

		w.Write([]byte("Login successful!"))
	})
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	return authorizationChan
}

func LoadOauthConfig(secretsFile string) *oauth2.Config {
	// Load the secrets file
	data, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		panic(err.Error())
	}

	// Read the secrets
	secrets := new(ClientSecrets)
	err = json.Unmarshal(data, secrets)
	if err != nil {
		panic(err.Error())
	}

	return &oauth2.Config{
		ClientID:     secrets.Web.ClientID,
		ClientSecret: secrets.Web.ClientSecret,
		RedirectURL:  "http://localhost:8080/oauth",
		Scopes: []string{
			"https://code.google.com/feeds/issues",
		},
		Endpoint: google.Endpoint,
	}
}

func LoadTokenClient(config *oauth2.Config, storageFile string) (*http.Client, error) {
	data, err := ioutil.ReadFile(storageFile)
	if err != nil {
		return nil, err
	}
	token := new(oauth2.Token)
	err = json.Unmarshal(data, token)
	if err != nil {
		return nil, err
	}
	client := config.Client(oauth2.NoContext, token)
	return client, nil
}

func StoreToken(token *oauth2.Token, storageFile string) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(storageFile, data, 0755)
	return err
}

func Authenticate(storageFile, secretsFile string) (*http.Client, error) {
	config := LoadOauthConfig(secretsFile)

	client, err := LoadTokenClient(config, storageFile)
	if err != nil {
		authorization := <-StartOauthHandler(config)
		err = StoreToken(authorization.Token, storageFile)
		if err != nil {
			return nil, err
		}
		return authorization.Client, nil
	}
	return client, nil
}
