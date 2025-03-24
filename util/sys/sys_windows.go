//go:build windows
// +build windows

package sys

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

func GetTCPCount() (int, error) {
	// Windows上使用netstat命令获取TCP连接数
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if strings.Contains(line, "TCP") &&
			(strings.Contains(line, "ESTABLISHED") ||
				strings.Contains(line, "LISTENING") ||
				strings.Contains(line, "TIME_WAIT")) {
			count++
		}
	}

	return count, nil
}

func GetUDPCount() (int, error) {
	// Windows上使用netstat命令获取UDP连接数
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if strings.Contains(line, "UDP") {
			count++
		}
	}

	return count, nil
}

func GetSystemInfo() string {
	return "Windows"
}

func GetPid(processName string) ([]int, error) {
	// 使用tasklist命令获取进程PID
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq "+processName, "/NH")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	pids := make([]int, 0)

	for _, line := range lines {
		if strings.Contains(line, processName) {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				pid, err := strconv.Atoi(fields[1])
				if err == nil {
					pids = append(pids, pid)
				}
			}
		}
	}

	if len(pids) == 0 {
		return nil, errors.New("process not found")
	}

	return pids, nil
}

func Kill(pid int) error {
	cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
	return cmd.Run()
}
