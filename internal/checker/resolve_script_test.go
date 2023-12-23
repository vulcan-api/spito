package checker

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestFetchRuleSet(t *testing.T) {
	ruleSetLocation := RuleSetLocation{}
	ruleSetLocation.New("https://github.com/Nasz-Elektryk/spito-ruleset/")

	err := ruleSetLocation.CreateDir()
	if err != nil {
		t.Fatal(err)
	}
	sets, err := GetAllDownloadedRuleSets()
	if err != nil {
		t.Fatal(err)
	}

	isRuleSetAlreadyDownloaded := slices.ContainsFunc(sets, func(s string) bool {
		return strings.Contains(s, ruleSetLocation.simpleUrl)
	})

	if isRuleSetAlreadyDownloaded {
		t.Log("!!! TEST SKIPPED !!!")
		t.Log("Test uses ruleset which you downloaded before running this test")
		t.Log("Delete ruleset called " + ruleSetLocation.simpleUrl + " if you want to run this test")

		t.SkipNow()
	}

	err = FetchRuleSet(&ruleSetLocation)
	if err != nil {
		t.Fatal(err)
	}

	err = os.RemoveAll(ruleSetLocation.GetRuleSetPath())
	if err != nil {
		t.Fatal(err)
	}
}
