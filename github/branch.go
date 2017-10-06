package github

import (
	"context"
	"log"
	"time"
)

// WaitUntilBranchisSynced checks if a specific branch is synced by Gerrit into Github.
// It is important that the branch exists at Github, because otherwise
// we won`t be able to create the merge request.
// Attention: This call is "kind of" blocking.
// It contains a for loop which ends only if the branch exists.
func (c GithubClient) waitUntilBranchisSynced(branchName string) error {
	ctx := context.Background()

	// Loop until branch is found on github and synced by Gerrit
	for {
		branch, _, err := c.Client.Repositories.GetBranch(ctx, c.Conf.Organisation, c.Conf.Repository, branchName)

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
