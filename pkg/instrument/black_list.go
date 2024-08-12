package instrument

import (
	"sdk-auto/pkg/process"
	"strings"
)

type Match interface {
	IsMatch(process *process.Details) bool
}

type Pattern interface {
	Match(value string) bool
}

type EqualPattern struct {
	value string
}

func (p *EqualPattern) Match(value string) bool {
	return value == p.value
}

type WildPattern struct {
	values []string
}

func (p *WildPattern) Match(value string) bool {
	subValue := value
	for _, matchValue := range p.values {
		if len(matchValue) > 0 {
			if index := strings.Index(subValue, matchValue); index != -1 {
				subValue = subValue[index:]
			} else {
				return false
			}

		}
	}
	return true
}

type CommMatch struct {
	patterns []Pattern
}

func NewCommMatch(blackList []string) *CommMatch {
	return &CommMatch{
		patterns: newPatterns(blackList),
	}
}

func (m *CommMatch) IsMatch(process *process.Details) bool {
	for _, pattern := range m.patterns {
		if pattern.Match(process.Comm) {
			return true
		}
	}
	return false
}

type CmdLineMatch struct {
	patterns []Pattern
}

func NewCmdLineMatch(blackList []string) *CmdLineMatch {
	return &CmdLineMatch{
		patterns: newPatterns(blackList),
	}
}

func (m *CmdLineMatch) IsMatch(process *process.Details) bool {
	for _, pattern := range m.patterns {
		if pattern.Match(process.CmdLine) {
			return true
		}
	}
	return false
}

func newPatterns(blackList []string) []Pattern {
	patterns := make([]Pattern, 0)
	for _, black := range blackList {
		patternArray := strings.Split(black, "*")
		if len(patternArray) == 1 {
			patterns = append(patterns, &EqualPattern{black})
		} else {
			patterns = append(patterns, &WildPattern{patternArray})
		}
	}
	return patterns
}
