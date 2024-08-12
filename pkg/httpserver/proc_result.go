package httpserver

import (
	"sdk-auto/pkg/inspectors"
	"sdk-auto/pkg/process"
)

type InstrumentProcess struct {
	Language string
	Pid      int
	Comm     string
	Cmd      string
	Exe      string
}

func NewInstrumentProcess(lang inspectors.ProgrammingLanguage, details *process.Details) InstrumentProcess {
	return InstrumentProcess{
		Language: string(lang),
		Pid:      details.ProcessID,
		Comm:     details.Comm,
		Cmd:      details.CmdLine,
		Exe:      details.ExeName,
	}
}
