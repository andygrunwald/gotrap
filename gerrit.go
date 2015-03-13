package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
func NewGerritInstance(c *gerritConfiguration) *GerritInstance {
	gerrit := &GerritInstance{
		URL:      c.URL,
		Username: c.Username,
		Password: c.Password,
		Template: strings.Join(c.Comment, "\n"),
	}

	return gerrit
}

// https://review.typo3.org/Documentation/rest-api-changes.html#set-review
func (g GerritInstance) postCommentOnChangeset(m *Message, vote int, msg string) {
	log.Printf("> Start posting review for %s (%s)", m.Change.URL, m.Patchset.Ref)

	changeID := m.Change.ID
	revisionID := m.Patchset.Revision
	urlToCall := fmt.Sprintf("%s/changes/%s/revisions/%s/review", g.getAPIUrl(), changeID, revisionID)

	log.Printf("> Calling %s", urlToCall)

	bodyStruct := &ReviewInput{
		Message: msg,
		Labels: map[string]int{
			// Code-Review
			"Verified": vote,
		},
	}

	body, _ := json.Marshal(bodyStruct)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlToCall, strings.NewReader(string(body)))
	req.SetBasicAuth(g.Username, g.Password)
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Println("> Call failed", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Printf("> Call success: %s", resp.Status)
	} else {
		log.Printf("> Call success, but the status code doesn`t match ~200: %s", resp.Status)
	}
}

func (g GerritInstance) getAPIUrl() string {
	host := strings.TrimRight(g.URL, "/") + "/a"
	return host
}

func (g GerritInstance) getChangeInformation(changeID string) (*ChangeInfo, error) {
	urlToCall := fmt.Sprintf("%s/changes/%s/?o=CURRENT_REVISION", g.getAPIUrl(), changeID)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", urlToCall, nil)
	req.SetBasicAuth(g.Username, g.Password)
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")

	log.Printf("> Getting details of change %s", changeID)

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Println("> Call failed", err)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Printf("> Change-details for change id \"%s\" received", changeID)

	} else {
		log.Printf("> Call success, but the status code doesn`t match ~200: %s", resp.Status)
		return nil, errors.New("Call success, but the status code doesn`t match ~200")
	}

	var change ChangeInfo

	respBody, err := ioutil.ReadAll(resp.Body)
	// Every Gerrit response starts with ")]}'"
	jsonBody := string(respBody)[4:]

	err = json.Unmarshal([]byte(jsonBody), &change)
	if err != nil {
		log.Fatal("> Reading JSON from Request failed", err)
		return nil, err
	}

	return &change, nil
}

func (g GerritInstance) isPatchsetTheCurrentPatchset(changeID string, patchsetNumber uint) (bool, error) {
	change, err := g.getChangeInformation(changeID)

	if err != nil {
		return false, err
	}

	if patchsetNumber == change.Revisions[change.CurrentRevision].Number {
		return true, nil
	}

	return false, nil
}
