// Package github provides all functionality to interact with github in point of view of gotrap.
// This includes getting the commit status of a merge request or create a new merge request.
package github

import (
	"github.com/andygrunwald/gotrap/config"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// GithubClient is the main data structure to interact with github.
// Client is to interact with github itself.
// Conf contains the github configuration.
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

// NewGithubClient will return a client to interact with Github.
// As an argument the github part of the configuration is necessary.
func NewGithubClient(conf *config.GithubConfiguration) *GithubClient {
	transport := &tokenSource{
		token: &oauth2.Token{AccessToken: conf.APIToken},
	}
	transportClient := oauth2.NewClient(oauth2.NoContext, transport)

	client := &GithubClient{
		Client: github.NewClient(transportClient),
		Conf:   conf,
	}

	return client
}
