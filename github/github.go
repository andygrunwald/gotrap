package github

import (
	"github.com/andygrunwald/gotrap/config"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
	"time"
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

// waitUntilCommitStatusIsAvailable checks if an external service (like TravisCI)
// already finished the process and reports back via the Github Commit Status API
func (c GithubClient) WaitUntilCommitStatusIsAvailable(pr github.PullRequest) (*github.CombinedStatus, error) {
	s := new(github.CombinedStatus)
	var err error

	// Wait one round before we start polling,
	// because in most cases the external service isn`t so fast
	time.Sleep(time.Duration(c.Conf.StatusPollingIntervall) * time.Second)

Loop:
	for {
		log.Printf("> Try to get commit status for %v/%v -> %v\n", c.Conf.Organisation, c.Conf.Repository, *pr.Head.Ref)
		s, _, err = c.Client.Repositories.GetCombinedStatus(c.Conf.Organisation, c.Conf.Repository, *pr.Head.Ref, nil)

		if err != nil {
			log.Printf("> Error during status fetch: %v\n", err)

		} else {
			log.Printf("> Commit status for %v/%v -> %v: %s", c.Conf.Organisation, c.Conf.Repository, *pr.Head.Ref, *s.State)
			switch *s.State {
			// Success if the latest status for all contexts is success
			case "success":
				break Loop

			// Pending if there are no statuses or a context is pending
			case "pending":
				time.Sleep(time.Duration(c.Conf.StatusPollingIntervall) * time.Second)

			// Failure if any of the contexts report as error or failure
			case "failure":
				break Loop
			}
		}
	}

	return s, nil
}
