package inspectors

import (
	"sdk-auto/pkg/process"
	"strings"
)

type JavaInspector struct{}

const processName = "java"

func (j *JavaInspector) Inspect(p *process.Details) (ProgrammingLanguage, bool) {
	if p.Comm == "sh" || p.Comm == "bash" {
		return "", false
	}
	if strings.Contains(p.ExeName, processName) || strings.Contains(p.CmdLine, processName) {
		return JavaProgrammingLanguage, true
	}

	return "", false
}
