package process

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Details struct {
	ProcessID int
	ExeName   string
	CmdLine   string
	Comm      string
}

func GetPidDetails(pid int) *Details {
	exeName := getExecName(pid)
	cmdLine := getCommandLine(pid)
	comm := GetComm(pid)

	return &Details{
		ProcessID: pid,
		ExeName:   exeName,
		CmdLine:   cmdLine,
		Comm:      comm,
	}
}

// The exe Symbolic Link: Inside each process's directory in /proc,
// there is a symbolic link named exe. This link points to the executable
// file that was used to start the process.
// For instance, if a process was started from /usr/bin/python,
// the exe symbolic link in that process's /proc directory will point to /usr/bin/python.
func getExecName(pid int) string {
	exeFileName := fmt.Sprintf("/proc/%d/exe", pid)
	exeName, err := os.Readlink(exeFileName)
	if err != nil {
		// Read link may fail if target process runs not as root
		return ""
	}
	return exeName
}

// reads the command line arguments of a Linux process from
// the cmdline file in the /proc filesystem and converts them into a string
func getCommandLine(pid int) string {
	cmdLineFileName := fmt.Sprintf("/proc/%d/cmdline", pid)
	fileContent, err := os.ReadFile(cmdLineFileName)
	if err != nil {
		// Ignore errors
		return ""
	} else {
		// \u0000替换为空格
		newByte := bytes.ReplaceAll([]byte(fileContent), []byte{0}, []byte{32})
		newByte = bytes.TrimSpace(newByte)
		return string(newByte)
	}
}

func GetComm(pid int) string {
	commFileName := fmt.Sprintf("/proc/%d/comm", pid)
	fileContent, err := os.ReadFile(commFileName)
	if err != nil {
		// Ignore errors
		return ""
	} else {
		comm := string(fileContent)
		// 移除换行符
		return strings.TrimSuffix(comm, "\n")
	}
}

func FindAllProcesses(predicate func(string) bool) ([]*Details, error) {
	dirs, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	var result []*Details
	for _, di := range dirs {

		if !di.IsDir() {
			continue
		}

		dirName := di.Name()

		pid, isProcessDirectory := isDirectoryPid(dirName)
		if !isProcessDirectory {
			continue
		}

		// predicate is optional, and can be used to filter the results
		// plus avoid doing unnecessary work (e.g. reading the command line and exe name)
		if predicate != nil && !predicate(dirName) {
			continue
		}

		details := GetPidDetails(pid)
		result = append(result, details)
	}

	return result, nil
}

func isDirectoryPid(procDirectoryName string) (int, bool) {
	if procDirectoryName[0] < '0' || procDirectoryName[0] > '9' {
		return 0, false
	}

	pid, err := strconv.Atoi(procDirectoryName)
	if err != nil {
		return 0, false
	}

	return pid, true
}
