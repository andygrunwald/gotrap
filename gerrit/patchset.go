package gerrit

func (g GerritInstance) IsPatchsetTheCurrentPatchset(changeID string, patchsetNumber uint) (bool, error) {
	change, err := g.getChangeInformation(changeID)

	if err != nil {
		return false, err
	}

	if patchsetNumber == change.Revisions[change.CurrentRevision].Number {
		return true, nil
	}

	return false, nil
}
