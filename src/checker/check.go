package checker

import (
	"regexp"
)

type Rule struct {
	url          string
	name         string
	isInProgress bool
}

type RulesHistory map[string]Rule

func (r RulesHistory) Contains(url string, name string) bool {
	_, ok := r[url+name]
	return ok
}

func (r RulesHistory) IsRuleInProgress(url string, name string) bool {
	val := r[url+name]
	return val.isInProgress
}

func (r RulesHistory) Push(url string, name string, isInProgress bool) {
	r[url+name] = Rule{
		url:          url,
		name:         name,
		isInProgress: isInProgress,
	}
}

func (r RulesHistory) SetProgress(url string, name string, isInProgress bool) {
	rule := r[url+name]
	rule.isInProgress = isInProgress
}

func CheckRuleScript(script string) (bool, error) {
 	rulesHistory := &RulesHistory{}
	return ExecuteLuaMain(script, rulesHistory)
}

func CheckRule(rulesHistory *RulesHistory, url string, name string) bool {
	const URL_REG_EXP = "https?:\\/\\/(www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)"
	isValid, err := regexp.MatchString(URL_REG_EXP, url)
	if err != nil || !isValid {
		panic("Rule url is invalid!")
	}
	if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	
	if rulesHistory.Contains(url, name) {
		if rulesHistory.IsRuleInProgress(url, name) {
			panic("ERROR: Dependencies creates infinity loop")
		} else {
			return true
		}
	}
	rulesHistory.Push(url, name, true)

	err, ruleSetPath, removeTempDir := FetchRuleSet(url)
	if err != nil {
		println(err.Error())
		panic("Failed to fetch rules from git repo: " + url)
	}
	defer removeTempDir()

	script, err := getScript(ruleSetPath, name)
	if err != nil {
		println(err.Error())
		panic("Failed to read script called: " + name + " in git repo: " + url)
	}

	rulesHistory.SetProgress(url, name, false)
	res, err := ExecuteLuaMain(script, rulesHistory)
	if err != nil {
		return false
	}
	return res
}
