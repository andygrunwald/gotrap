package gerrit

import (
	"log"
)

func (g GerritInstance) IsPatchsetTheCurrentPatchset(changeID string, patchsetNumber uint) (bool, error) {
	log.Printf("> Getting details of change %s", changeID)
	change, err := g.getChangeInformation(changeID)

	if err != nil {
		return false, err
	}

	if patchsetNumber == change.Revisions[change.CurrentRevision].Number {
		return true, nil
	}

	return false, nil
}
