package shared

type Rule struct {
	Url          string
	NameOrScript string
	IsScript     bool
	isInProgress bool
}

func (r Rule) GetIdentifier() string {
	return r.Url + r.NameOrScript
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

func (r RulesHistory) Push(url string, nameOrScript string, isInProgress, isScript bool) {
	r[url+nameOrScript] = Rule{
		Url:          url,
		NameOrScript: nameOrScript,
		IsScript:     isScript,
		isInProgress: isInProgress,
	}
}

func (r RulesHistory) SetProgress(url string, nameOrScript string, isInProgress bool) {
	rule := r[url+nameOrScript]
	rule.isInProgress = isInProgress
}
