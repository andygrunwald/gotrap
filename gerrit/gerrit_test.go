package gerrit

import (
	"strings"
	"testing"
)

func TestGetAPIUrlIsNotEmpty(t *testing.T) {
	g := &GerritInstance{
		URL: "https://review.typo3.org/",
	}

	if url := g.getAPIUrl(); len(url) == 0 {
		t.Fail()
	}
}

func TestGetAPIUrlContainsAuthPrefix(t *testing.T) {
	g := &GerritInstance{
		URL: "https://review.typo3.org/",
	}

	if url := g.getAPIUrl(); strings.HasSuffix(url, "/a") == false {
		t.Fail()
	}
}
