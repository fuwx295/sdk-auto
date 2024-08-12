package inspectors

import (
	"debug/buildinfo"
	"fmt"
	"sdk-auto/pkg/process"
)

type GolangInspector struct{}

func (g *GolangInspector) Inspect(p *process.Details) (ProgrammingLanguage, bool) {
	if p.Comm == "" {
		// comm作为服务名.
		return "", false
	}
	file := fmt.Sprintf("/proc/%d/exe", p.ProcessID)
	_, err := buildinfo.ReadFile(file)
	if err != nil {
		return "", false
	}

	return GoProgrammingLanguage, true
}
