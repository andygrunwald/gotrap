package stream

import (
	"encoding/json"
	"fmt"
	"github.com/andygrunwald/gotrap/config"
	"github.com/andygrunwald/gotrap/gerrit"
	"github.com/andygrunwald/gotrap/github"
	"github.com/streadway/amqp"
	"log"
	"strings"
)

func handleNewMessage(g github.GithubClient, gerritClient gerrit.GerritInstance, c *config.Configuration, event amqp.Delivery) {
	// Acknowledge message if we get this
	// We do this, because at the end of proceeding we might lost the connection to AMQP server
	// I know this is wrong, but currently the reconnection does not work correctly :(
	// TODO: Fix this later and Ack message if its done
	event.Ack(false)

	var change gerrit.Message

	err := json.Unmarshal(event.Body, &change)
	if err != nil {
		log.Fatal(err)
	}

	// Stream events are documented
	// See https://git.eclipse.org/r/Documentation/cmd-stream-events.html
	switch change.Type {
	case "patchset-created":
		log.Printf("> New patchset-created message incoming for ref \"%s\" in \"%s\" (\"%s\")", change.Patchset.Ref, change.Change.Project, change.Change.URL)

		if change.Change.Project != "Packages/TYPO3.CMS" {
			log.Printf("> Project \"%s\" currently not supported. Only \"%s\"", change.Change.Project, "Packages/TYPO3.CMS")
			return
		}

		// TODO: Get current revision number
		// If this revision / patchset number is not the current number
		// we will skip this patchset-created request, because
		// why should we create a pull request for an old patchset?
		// The current patchset will be delivered later as message.
		// So we won`t skip this changeset
		// https://review.typo3.org/a/changes/I640486e9f32da6ac1eba05e3c38d15a0aba41055/?o=CURRENT_REVISION
		if currentPatchset, _ := gerritClient.IsPatchsetTheCurrentPatchset(change.Change.ID, change.Patchset.Number); currentPatchset == false {
			log.Printf("> Patchset skipped, because it is not the current one (Ref: %s)", change.Patchset.Ref)
			return
		}

		// Create the pull request
		pullRequest, err := g.CreatePullRequestForPatchset(&change)
		if err != nil {
			// TODO: I don`t have an idea what to do if i fail to create a PR
			log.Println("> Error during creating new pull request", err)
			return
		}

		log.Printf("> New pull request created: %s", *pullRequest.HTMLURL)

		// Poll travis ci and wait until the PR got a status
		s, _ := g.WaitUntilCommitStatusIsAvailable(*pullRequest)

		var vote int
		switch *s.State {
		// Success if the latest status for all contexts is success
		case "success":
			vote = 0

		// Failure if any of the contexts report as error or failure
		case "failure":
			vote = -1
		}

		// Generate detail information about status
		var statusDetails []string
		for _, repoStatus := range s.Statuses {
			if *repoStatus.State != "success" {
				continue
			}

			statusDetails = append(statusDetails, "Service: "+*repoStatus.Context)
			statusDetails = append(statusDetails, "Description: "+*repoStatus.Description)
			statusDetails = append(statusDetails, "URL: "+*repoStatus.TargetURL)
			statusDetails = append(statusDetails, "\n")
		}

		// Build template for Gerrit vote action
		msg := gerritClient.Template
		msg = strings.Replace(msg, "%state%", *s.State, 1)
		msg = strings.Replace(msg, "%status%", strings.Join(statusDetails, "\n\n"), 1)
		msg = strings.Replace(msg, "%pr%", *pullRequest.HTMLURL, 1)

		// Post Command + Vote on Changeset
		gerritClient.PostCommentOnChangeset(&change, vote, msg)

		msg = "This PR will be closed, because the tests results were reported back to Gerrit. "
		msg += fmt.Sprintf("See [%s](%s) for details.", change.Change.Subject, change.Change.URL)

		// TODO Add Details to Github Pull Request, before closing
		// This does not work currently. I don`t have a clue why. Check this later ;)
		g.AddCommentToPullRequest(pullRequest, msg)

		g.ClosePullRequest(pullRequest)

	case "change-abandoned":
		// We have to close all PR`s
		// change-restored
		// change-merged
		// ref-updated
		// ref-replicated
		// ref-replication-done
		// comment-added
		// topic-changed
		// ....
	default:
		log.Printf("> Skipped AMQP message (uncovered message type: %s)\n", change.Type)
	}

	return
}
