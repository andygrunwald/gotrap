package stream

import (
	"bytes"
	"fmt"
	"github.com/andygrunwald/gotrap/config"
	"github.com/andygrunwald/gotrap/gerrit"
	"github.com/andygrunwald/gotrap/github"
	"log"
	"regexp"
	"text/template"
)

type Gotrap struct {
	githubClient github.GithubClient
	gerritClient gerrit.GerritInstance
	config       *config.Configuration
	Message      gerrit.Message
}

func (trap *Gotrap) TakeAction() {
	// Stream events are documented
	// See https://git.eclipse.org/r/Documentation/cmd-stream-events.html
	switch trap.Message.Type {
	case "patchset-created":
		log.Printf("> New patchset-created message incoming for ref \"%s\" in \"%s\" (%s)", trap.Message.Patchset.Ref, trap.Message.Change.Project, trap.Message.Change.URL)

		// Check if Project is configured
		if _, err := trap.IsProjectConfigured(trap.Message.Change.Project); err != nil {
			log.Printf("> %s", err)
			return
		}

		// Check if branch is configured
		if _, err := trap.IsBranchConfigured(trap.Message.Change.Project, trap.Message.Change.Branch); err != nil {
			log.Printf("> %s", err)
			return
		}

		log.Printf("> Getting details of change %s", trap.Message.Change.ID)
		gerritChangeSet, err := trap.gerritClient.GetChangeInformation(trap.Message.Change.ID)
		if err != nil {
			log.Printf("> Error getting details of change %s creating new pull request: %s", trap.Message.Change.ID, err)
			return
		}

		// Check if the status of the changeset is NEW and not
		// SUBMITTED, MERGED, ABANDONED or DRAFT
		// We only accept NEW changesets
		if gerritChangeSet.Status != "NEW" {
			log.Printf("> Changeset skipped, because status is \"%s\" and not \"NEW\"", gerritChangeSet.Status)
			return
		}

		// TODO: Get current revision number
		// If this revision / patchset number is not the current number
		// we will skip this patchset-created request, because
		// why should we create a pull request for an old patchset?
		// The current patchset will be delivered later as message.
		// So we won`t skip this changeset
		// https://review.typo3.org/a/changes/I640486e9f32da6ac1eba05e3c38d15a0aba41055/?o=CURRENT_REVISION
		if currentPatchset, _ := trap.gerritClient.IsPatchsetTheCurrentPatchset(gerritChangeSet, trap.Message.Patchset.Number); currentPatchset == false {
			log.Printf("> Patchset skipped, because it is not the current one (Ref: %s of %s)", trap.Message.Patchset.Ref, trap.Message.Change.URL)
			return
		}

		// Check if change subject is excluded
		if res, matchedPattern := trap.IsSubjectExcludedByPattern(trap.Message.Change.Subject); res == true {
			log.Printf("> Subject \"%s\" excluded by pattern \"%s\"", trap.Message.Change.Subject, matchedPattern)
			return
		}

		// Create the pull request
		pullRequest, err := trap.githubClient.CreatePullRequestForPatchset(&trap.Message)
		if err != nil {
			// TODO: I don`t have an idea what to do if i fail to create a PR
			log.Println("> Error during creating new pull request", err)
			return
		}

		log.Printf("> New pull request created: %s", *pullRequest.HTMLURL)

		// Poll travis ci and wait until the PR got a status
		s, _ := trap.githubClient.WaitUntilCommitStatusIsAvailable(*pullRequest)

		// Build a combined data structure for templating
		gotrapResult := github.PullRequest{
			PullRequest:    pullRequest,
			CombinedStatus: s,
		}

		var vote int
		// We only take care about every status except of "pending"
		switch *s.State {
		// Success if the latest status for all contexts is success
		case "success":
			vote = 0

		// Error is the 3rd party service fails
		case "error":
			vote = 0

		// Failure if any of the contexts report as error or failure
		case "failure":
			vote = -1
		}

		// Build message to post results back to Gerrit
		statusDetailsBuffer := new(bytes.Buffer)
		var statusDetailsTemplate = template.Must(template.New("status-details").Parse(trap.gerritClient.Template))
		err = statusDetailsTemplate.Execute(statusDetailsBuffer, gotrapResult)
		if err != nil {
			log.Println("> Error during prepare the status detail message", err)
			return
		}

		// Post Command + Vote on Changeset
		trap.gerritClient.PostCommentOnChangeset(&trap.Message, vote, statusDetailsBuffer.String())

		// TODO: Make text configurable
		msg := "This PR will be closed, because the tests results were reported back to Gerrit. See [{{.Message.Change.Subject}}]({{.Message.Change.URL}}) for details."

		// Build message to close the Pull Request
		closeMsgBuffer := new(bytes.Buffer)
		var closeMsgTemplate = template.Must(template.New("pull-request-close-message").Parse(msg))
		err = closeMsgTemplate.Execute(closeMsgBuffer, *trap)
		if err != nil {
			log.Println("> Error during prepare the pull request close message", err)
			return
		}

		_, err = trap.githubClient.AddCommentToPullRequest(pullRequest, closeMsgBuffer.String())
		if err != nil {
			log.Printf("> Error during adding a comment to a pull request %s: %s", *pullRequest.HTMLURL, err)
		} else {
			log.Printf("> Comment added to pull request: %s", *pullRequest.HTMLURL)
		}

		_, err = trap.githubClient.ClosePullRequest(pullRequest)
		if err != nil {
			log.Printf("> Error during closing a pull request %s: %s", *pullRequest.HTMLURL, err)
		} else {
			log.Printf("> Pull request closed: %s", *pullRequest.HTMLURL)
		}

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
		log.Printf("> Skipped AMQP message (uncovered message type: %s)\n", trap.Message.Type)
	}

	return
}

func (trap *Gotrap) IsProjectConfigured(project string) (bool, error) {
	if _, ok := trap.config.Gerrit.Projects[project]; !ok {
		return false, fmt.Errorf("Project \"%s\" is not configured", project)
	}

	return true, nil
}

func (trap *Gotrap) IsBranchConfigured(project, branch string) (bool, error) {
	// If no branch is configured, we assume that all branches should be covered
	if len(trap.config.Gerrit.Projects[project]) == 0 {
		return true, nil
	}

	// If the branch exists and is configured as true, then it is valid configured
	if val, ok := trap.config.Gerrit.Projects[project][branch]; ok && val == true {
		return true, nil
	}

	return false, fmt.Errorf("Branch \"%s\" not configured for project \"%s\"", branch, project)
}

func (trap *Gotrap) IsSubjectExcludedByPattern(subject string) (bool, string) {
	for _, pattern := range trap.config.Gerrit.ExcludePattern {
		if matched, err := regexp.MatchString(pattern, subject); err == nil && matched == true {
			return matched, pattern
		}
	}

	return false, ""
}
