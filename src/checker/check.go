package checker

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

func CheckRule(rulesHistory *RulesHistory, identifier string, name string) bool {
	ruleSetLocation := RuleSetLocation{}
	ruleSetLocation.new(identifier)
	simpleUrl := ruleSetLocation.simpleUrl

	if rulesHistory.Contains(simpleUrl, name) {
		if rulesHistory.IsRuleInProgress(simpleUrl, name) {
			panic("ERROR: Dependencies creates infinity loop")
		} else {
			return true
		}
	}
	rulesHistory.Push(simpleUrl, name, true)

	err := FetchRuleSet(&ruleSetLocation)
	if err != nil {
		panic("Failed to fetch rules from git: " + ruleSetLocation.getFullUrl() + "\n" + err.Error())
	}

	script, err := getScript(ruleSetLocation, name)
	if err != nil {
		println(err.Error())
		panic("Failed to read script called: " + name + " in git: " + ruleSetLocation.getFullUrl())
	}

	rulesHistory.SetProgress(simpleUrl, name, false)
	res, err := ExecuteLuaMain(script, rulesHistory)
	if err != nil {
		return false
	}
	return res
}
