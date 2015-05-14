package github

import (
	"errors"
	"github.com/andygrunwald/gotrap/gerrit"
	"github.com/google/go-github/github"
	"strings"
)

// createPullRequestForPatchset will create a new Pull Request at Github
// All information (like base and target branch) are received by the message by Gerrit
func (c GithubClient) CreatePullRequestForPatchset(m *gerrit.Message) (*github.PullRequest, error) {

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

func (c GithubClient) AddCommentToPullRequest(pr *github.PullRequest, message string) (bool, error) {

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

func (c GithubClient) ClosePullRequest(pr *github.PullRequest) (bool, error) {
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
