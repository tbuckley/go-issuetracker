package googauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type webSecrets struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}
type clientSecrets struct {
	Web webSecrets `json:"web"`
}

func exchangeCode(config *oauth2.Config, code string) (*oauth2.Token, *http.Client, error) {
	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, nil, err
	}
	client := config.Client(oauth2.NoContext, token)
	return token, client, nil
}

func getOauthURL(config *oauth2.Config) string {
	return config.AuthCodeURL("foobar", oauth2.AccessTypeOffline)
}

type authorization struct {
	Client *http.Client
	Token  *oauth2.Token
}

func startOauthHandler(config *oauth2.Config) chan *authorization {
	authorizationChan := make(chan *authorization)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := getOauthURL(config)
		w.Write([]byte("<a href=\"" + url + "\">Click here</a>"))
	})
	http.HandleFunc("/oauth", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, client, err := exchangeCode(config, code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		authorizationChan <- &authorization{
			Client: client,
			Token:  token,
		}

		w.Write([]byte("Login successful! You may now close this page."))
	})
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	go func() {
		time.Sleep(1 * time.Second)
		err := open.Run("http://localhost:8080")
		if err != nil {
			log.Fatal(err)
		}
	}()

	return authorizationChan
}

func loadOauthConfig(secretsFile string) *oauth2.Config {
	// Load the secrets file
	data, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		panic(err.Error())
	}

	// Read the secrets
	secrets := new(clientSecrets)
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

func loadTokenClient(config *oauth2.Config, storageFile string) (*http.Client, error) {
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

func storeToken(token *oauth2.Token, storageFile string) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	log.Printf("Writing to file %v", storageFile)
	return ioutil.WriteFile(storageFile, data, 0755)
}

func Authenticate(storageFile, secretsFile string) (*http.Client, error) {
	config := loadOauthConfig(secretsFile)

	client, err := loadTokenClient(config, storageFile)
	if err != nil {
		authChan := startOauthHandler(config)
		fmt.Println("Visit http://localhost:8080 to log in")
		authorization := <-authChan
		err = storeToken(authorization.Token, storageFile)
		if err != nil {
			return nil, err
		}
		return authorization.Client, nil
	}
	return client, nil
}
