package github

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/github"
)

// waitUntilCommitStatusIsAvailable checks if an external service (like TravisCI)
// already finished the process and reports back via the Github Commit Status API
func (c GithubClient) WaitUntilCommitStatusIsAvailable(pr github.PullRequest) (*github.CombinedStatus, error) {
	s := new(github.CombinedStatus)
	var err error

	// Wait one round before we start polling,
	// because in most cases the external service isn`t so fast
	time.Sleep(time.Duration(c.Conf.StatusPollingIntervall) * time.Second)
	ctx := context.Background()

Loop:
	for {
		log.Printf("> Try to get commit status for %v/%v -> %v\n", c.Conf.Organisation, c.Conf.Repository, *pr.Head.Ref)
		s, _, err = c.Client.Repositories.GetCombinedStatus(ctx, c.Conf.Organisation, c.Conf.Repository, *pr.Head.Ref, nil)

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
			case "error":
				break Loop

			// Failure if any of the contexts report as error or failure
			case "failure":
				break Loop
			}
		}
	}

	return s, nil
}
