package gerrit

import (
	"strings"
	"testing"
)

func TestGetAPIUrlIsNotEmpty(t *testing.T) {
	g := &GerritInstance{
		URL: "https://review.typo3.org/",
	}

	if url := g.getAPIUrl(false); len(url) == 0 {
		t.Fail()
	}
}

func TestGetAPIUrlContainsAuthPrefix(t *testing.T) {
	g := &GerritInstance{
		URL: "https://review.typo3.org/",
	}

	if url := g.getAPIUrl(true); strings.HasSuffix(url, "/a") == false {
		t.Fail()
	}
}
