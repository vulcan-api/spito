package shared

type Rule struct {
	Url          string
	Name         string
	isInProgress bool
}

func (r Rule) GetIdentifier() string {
	return r.Url + r.Name
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
		Url:          url,
		Name:         name,
		isInProgress: isInProgress,
	}
}

func (r RulesHistory) SetProgress(url string, name string, isInProgress bool) {
	rule := r[url+name]
	rule.isInProgress = isInProgress
}
