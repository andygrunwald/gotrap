package main

import (
	"code.google.com/p/goauth2/oauth"
	"errors"
	"github.com/google/go-github/github"
	"log"
	"strings"
	"time"
)

// GithubClient offers the functionality to communicate with Github
type GithubClient struct {
	Client *github.Client
	Conf   *githubConfiguration
}

// NewGithubClient will return a new Github client
func NewGithubClient(conf *githubConfiguration) *GithubClient {
	// Simple OAuth transport
	transport := &oauth.Transport{
		Token: &oauth.Token{AccessToken: conf.APIToken},
	}

	c := github.NewClient(transport.Client())

	client := &GithubClient{
		Client: c,
		Conf:   conf,
	}

	return client
}

// WaitUntilBranchisSynced checks if a specific branch is synced by Gerrit into Github.
// It is important that the branch exists at Github, because otherwise
// we won`t be able to create the merge request.
// Attention: This call is "kind of" blocking.
// It contains a for loop which ends only if the branch exists.
func (c GithubClient) waitUntilBranchisSynced(branchName string) error {

	// Loop until branch is found on github and synced by Gerrit
	for {
		branch, _, err := c.Client.Repositories.GetBranch(c.Conf.Organisation, c.Conf.Repository, branchName)

		// A typical error can be
		// GET https://api.github.com/repos/... 404 Branch not found []
		// We will log this and keep polling, until this is synced
		if err != nil {
			// TODO Max loops
			log.Printf("> Wait until branch \"%s\" is synced to %s/%s: %v", branchName, c.Conf.Organisation, c.Conf.Repository, err)

		} else {
			log.Printf("> Branch \"%s\" found on %s/%s", *branch.Name, c.Conf.Organisation, c.Conf.Repository)
			break
		}

		time.Sleep(time.Duration(c.Conf.BranchPollingIntervall) * time.Second)
	}

	return nil
}

// waitUntilCommitStatusIsAvailable checks if an external service (like TravisCI)
// already finished the process and reports back via the Github Commit Status API
func (c GithubClient) waitUntilCommitStatusIsAvailable(pr github.PullRequest) (*github.CombinedStatus, error) {
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

// createPullRequestForPatchset will create a new Pull Request at Github
// All information (like base and target branch) are received by the message by Gerrit
func (c GithubClient) createPullRequestForPatchset(m *Message) (*github.PullRequest, error) {

	// Remove "refs/" from the patchset reference,
	// because if this patchset is synced to Github
	// the branch is named without "refs/" as prefix.
	baseRef := m.Patchset.Ref
	if strings.HasPrefix(baseRef, "refs/") {
		baseRef = baseRef[5:]
	}

	// Start polling until the branch is synced
	// We have to wait, because after this we are able to continue
	err := c.waitUntilBranchisSynced(baseRef)

	if err != nil {
		return nil, errors.New("Max loop reached for branch")
	}

	title := c.Conf.PRTemplate.Title
	title = strings.Replace(title, "%title%", m.Change.Subject, 1)

	body := strings.Join(c.Conf.PRTemplate.Body, "\n")
	body = strings.Replace(body, "%commit-msg%", m.Change.CommitMessage, 1)
	body = strings.Replace(body, "%url%", m.Change.URL, 1)

	pr := &github.NewPullRequest{
		Title: &title,
		Head:  &baseRef,
		Base:  &m.Change.Branch,
		Body:  &body,
	}

	// Do the pull request itself
	prResult, resp, err := c.Client.PullRequests.Create(c.Conf.Organisation, c.Conf.Repository, pr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return prResult, nil
}

func (c GithubClient) addCommentToPullRequest(pr *github.PullRequest, message string) (bool, error) {

	comment := &github.PullRequestComment{
		Body: &message,
	}
	_, resp, err := c.Client.PullRequests.CreateComment(c.Conf.Organisation, c.Conf.Repository, *pr.Number, comment)
	// func (s *PullRequestsService) CreateComment(owner string, repo string, number int, comment *PullRequestComment) (*PullRequestComment, *Response, error)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

func (c GithubClient) closePullRequest(pr *github.PullRequest) (bool, error) {
	// func (s *PullRequestsService) Edit(owner string, repo string, number int, pull *PullRequest) (*PullRequest, *Response, error)
	state := "closed"
	updatePr := &github.PullRequest{
		State: &state,
	}

	_, resp, err := c.Client.PullRequests.Edit(c.Conf.Organisation, c.Conf.Repository, *pr.Number, updatePr)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}
