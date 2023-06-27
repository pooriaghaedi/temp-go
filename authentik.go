package main

import (
	"net/http"

	"golang.org/x/oauth2"
)

// Provider is the implementation of `goth.Provider` for accessing Authentik.
type Provider struct {
	ClientKey    string
	Secret       string
	CallbackURL  string
	HTTPClient   *http.Client
	config       *oauth2.Config
	providerName string
}

func New(clientKey, secret, callbackURL string, scopes ...string) *Provider {
	p := &Provider{
		ClientKey:   clientKey,
		Secret:      secret,
		CallbackURL: callbackURL,
	}
	p.config = newConfig(p, scopes)
	return p
}

func newConfig(provider *Provider, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     provider.ClientKey,
		ClientSecret: provider.Secret,
		RedirectURL:  provider.CallbackURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://auth.local/application/o/authorize/",
			TokenURL: "http://auth.local/application/o/token/",
		},
	}
}

// func (p *Provider) FetchUser(session goth.Session) (goth.User, error) {
// 	sess := session.(*Session)
// 	user := goth.User{
// 		AccessToken: sess.AccessToken,
// 		Provider:    p.Name(),
// 	}
// 	// Make a request to fetch user data
// 	// Parse the user data to the goth.User struct
// 	// return user, nil
// }
