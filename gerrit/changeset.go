package gerrit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func (g GerritInstance) getChangeInformation(changeID string) (*ChangeInfo, error) {
	urlToCall := fmt.Sprintf("%s/changes/%s/?o=CURRENT_REVISION", g.getAPIUrl(), changeID)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", urlToCall, nil)
	req.SetBasicAuth(g.Username, g.Password)
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
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
		return nil, err
	}

	return &change, nil
}

// https://review.typo3.org/Documentation/rest-api-changes.html#set-review
func (g GerritInstance) PostCommentOnChangeset(m *Message, vote int, msg string) {
	log.Printf("> Start posting review for %s (%s)", m.Change.URL, m.Patchset.Ref)

	changeID := m.Change.ID
	revisionID := m.Patchset.Revision
	urlToCall := fmt.Sprintf("%s/changes/%s/revisions/%s/review", g.getAPIUrl(true), changeID, revisionID)

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
