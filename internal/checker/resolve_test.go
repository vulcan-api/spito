package checker

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestFetchRuleSet(t *testing.T) {
	ruleSetLocation := NewRulesetLocation("https://github.com/avorty/spito-ruleset/")

	err := ruleSetLocation.CreateDir()
	if err != nil {
		t.Fatal(err)
	}
	sets, err := GetAllDownloadedRuleSets()
	if err != nil {
		t.Fatal(err)
	}

	isRuleSetAlreadyDownloaded := slices.ContainsFunc(sets, func(s string) bool {
		return strings.Contains(s, ruleSetLocation.GetIdentifier())
	})

	if isRuleSetAlreadyDownloaded {
		t.Log("!!! TEST SKIPPED !!!")
		t.Log("Test uses ruleset which you downloaded before running this test")
		t.Log("Delete ruleset called " + ruleSetLocation.GetIdentifier() + " if you want to run this test")

		t.SkipNow()
	}

	err = FetchRuleset(&ruleSetLocation)
	if err != nil {
		t.Fatal(err)
	}

	err = os.RemoveAll(ruleSetLocation.GetRulesetPath())
	if err != nil {
		t.Fatal(err)
	}
}
