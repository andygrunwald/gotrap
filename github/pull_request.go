package github

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"text/template"

	"github.com/andygrunwald/gotrap/gerrit"
	"github.com/google/go-github/github"
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

	// Build title for Pull Request
	titleBuffer := new(bytes.Buffer)
	var titleTemplate = template.Must(template.New("pull-request-title").Parse(c.Conf.PRTemplate.Title))
	err = titleTemplate.Execute(titleBuffer, m)
	if err != nil {
		return nil, err
	}
	title := titleBuffer.String()

	// Build body for Pull Request
	bodyString := strings.Join(c.Conf.PRTemplate.Body, "\n")
	bodyBuffer := new(bytes.Buffer)
	var bodyTemplate = template.Must(template.New("pull-request-body").Parse(bodyString))
	err = bodyTemplate.Execute(bodyBuffer, m)
	if err != nil {
		return nil, err
	}
	body := bodyBuffer.String()

	// Create the pull request itself
	pr := &github.NewPullRequest{
		Title: &title,
		Head:  &baseRef,
		Base:  &m.Change.Branch,
		Body:  &body,
	}
	ctx := context.Background()
	prResult, resp, err := c.Client.PullRequests.Create(ctx, c.Conf.Organisation, c.Conf.Repository, pr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return prResult, nil
}

func (c GithubClient) AddCommentToPullRequest(pr *github.PullRequest, message string) (bool, error) {
	comment := &github.IssueComment{
		Body: &message,
	}

	ctx := context.Background()
	_, resp, err := c.Client.Issues.CreateComment(ctx, c.Conf.Organisation, c.Conf.Repository, *pr.Number, comment)

	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

func (c GithubClient) ClosePullRequest(pr *github.PullRequest) (bool, error) {
	state := "closed"
	updatePr := &github.PullRequest{
		State: &state,
	}

	ctx := context.Background()
	_, resp, err := c.Client.PullRequests.Edit(ctx, c.Conf.Organisation, c.Conf.Repository, *pr.Number, updatePr)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}
