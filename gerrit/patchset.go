package gerrit

func (g GerritInstance) IsPatchsetTheCurrentPatchset(change *ChangeInfo, patchsetNumber uint) (bool, error) {
	if patchsetNumber == change.Revisions[change.CurrentRevision].Number {
		return true, nil
	}

	return false, nil
}
