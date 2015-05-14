package github

import (
	"github.com/andygrunwald/gotrap/config"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// GithubClient offers the functionality to communicate with Github
type GithubClient struct {
	Client *github.Client
	Conf   *config.GithubConfiguration
}

// tokenSource is an oauth2.TokenSource which returns a static access token
type tokenSource struct {
	token *oauth2.Token
}

// Token implements the oauth2.TokenSource interface
func (t *tokenSource) Token() (*oauth2.Token, error) {
	return t.token, nil
}

// NewGithubClient will return a new Github client
func NewGithubClient(conf *config.GithubConfiguration) *GithubClient {

	transport := &tokenSource{
		token: &oauth2.Token{AccessToken: conf.APIToken},
	}

	transportClient := oauth2.NewClient(oauth2.NoContext, transport)

	c := github.NewClient(transportClient)

	client := &GithubClient{
		Client: c,
		Conf:   conf,
	}

	return client
}
