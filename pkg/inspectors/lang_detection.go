package inspectors

import "sdk-auto/pkg/process"

type ProgrammingLanguage string

const (
	JavaProgrammingLanguage ProgrammingLanguage = "java"
	GoProgrammingLanguage   ProgrammingLanguage = "go"
)

type inspector interface {
	Inspect(process *process.Details) (ProgrammingLanguage, bool)
}

var inspectorsList = []inspector{
	&GolangInspector{},
	&JavaInspector{},
}

// DetectLanguage returns the detected language for the process or nil if the language could not be detected
func DetectLanguage(process *process.Details) *ProgrammingLanguage {
	for _, i := range inspectorsList {
		language, detected := i.Inspect(process)
		if detected {
			return &language
		}
	}

	return nil
}
