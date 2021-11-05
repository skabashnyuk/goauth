package main

import (
	"fmt"
	"github.com/google/go-github/github"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	"log"
	"net/http"
)

type OAuth2Config struct {
	ClientID     string `required:"true" envconfig:"CLIENT_ID"`
	ClientSecret string `required:"true" envconfig:"CLIENT_SECRET"`
}

var oauthStateString = "thisshouldberandom"
var oauthConf oauth2.Config

const htmlIndex = `<html><body>
Logged in with <a href="/login">GitHub</a>
</body></html>
`

// /
func handleMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlIndex))
}

// /login
func handleGitHubLogin( w http.ResponseWriter, r *http.Request) {
	url := oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// /github_oauth_cb. Called by github after authorization is granted
func handleGitHubCallback( w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	oauthClient := oauthConf.Client(oauth2.NoContext, token)
	client := github.NewClient(oauthClient)
	user, _, err := client.Users.Get(oauth2.NoContext, "")
	if err != nil {
		fmt.Printf("client.Users.Get() faled with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	fmt.Printf("Logged in as GitHub user: %s\n", *user.Login)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func main() {
	var config OAuth2Config
	err := envconfig.Process("myapp", &config)
	if err != nil {
		log.Fatal(err.Error())
	}
	format := "ClientID: %s\nClientSecret: %s\n"
	_, err = fmt.Printf(format, config.ClientID, config.ClientSecret)
	if err != nil {
		log.Fatal(err.Error())
	}
	 oauthConf = oauth2.Config{

	 	ClientID: config.ClientID,
	 	ClientSecret: config.ClientSecret,
		// select level of access you want https://developer.github.com/v3/oauth/#scopes
		Scopes:   []string{"user:email", "repo"},
		Endpoint: githuboauth.Endpoint,
	}

	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleGitHubLogin)
	http.HandleFunc("/github_oauth_cb", handleGitHubCallback)
	fmt.Print("Started running on http://127.0.0.1:7000\n")
	fmt.Println(http.ListenAndServe(":7000", nil))
}
