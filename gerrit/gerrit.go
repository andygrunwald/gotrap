package gerrit

import (
	"github.com/andygrunwald/gotrap/config"
	"strings"
)

type GerritInstance struct {
	URL      string
	Username string
	Password string
	Template string
}

// https://review.typo3.org/Documentation/rest-api-changes.html#review-input
type ReviewInput struct {
	Message string         `json:"message"`
	Labels  map[string]int `json:"labels"`
}

// @link https://review.typo3.org/Documentation/json.html#patchSet
type Patchset struct {
	Ref      string `json:"ref"`
	Revision string `json:"revision"`
	Number   uint   `json:"number,string"`
}

// @link https://review.typo3.org/Documentation/json.html#change
type Change struct {
	Project       string `json:"project"`
	Branch        string `json:"branch"`
	ID            string `json:"id"`
	Subject       string `json:"subject"`
	CommitMessage string `json:"commitMessage"`
	URL           string `json:"url"`
}

// @link https://review.typo3.org/Documentation/cmd-stream-events.html#events
type Message struct {
	Type     string   `json:"type"`
	Change   Change   `json:"change"`
	Patchset Patchset `json:"patchSet"`
}

// @link https://review.typo3.org/Documentation/rest-api-changes.html#change-info
type ChangeInfo struct {
	CurrentRevision string `json:"current_revision"`
	Revisions       map[string]RevisionInfo
}

// @link https://review.typo3.org/Documentation/rest-api-changes.html#revision-info
type RevisionInfo struct {
	Number uint `json:"_number"`
}

// NewGerritInstance returns a new Gerrit instance
func NewGerritClient(c *config.GerritConfiguration) *GerritInstance {
	gerrit := &GerritInstance{
		URL:      c.URL,
		Username: c.Username,
		Password: c.Password,
		Template: strings.Join(c.Comment, "\n"),
	}

	return gerrit
}

func (g GerritInstance) getAPIUrl() string {
	host := strings.TrimRight(g.URL, "/") + "/a"
	return host
}
