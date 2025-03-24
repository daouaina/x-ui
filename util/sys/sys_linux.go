//go:build linux
// +build linux

package sys

import (
	"bytes"
	"io"
	"os"
	"strings"
)

func getLinesNum(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	sum := 0
	buf := make([]byte, 8192)
	for {
		n, err := file.Read(buf)

		var buffPosition int
		for {
			i := bytes.IndexByte(buf[buffPosition:], '\n')
			if i < 0 || n == buffPosition {
				break
			}
			buffPosition += i + 1
			sum++
		}

		if err == io.EOF {
			return sum, nil
		} else if err != nil {
			return sum, err
		}
	}
}

func GetTCPCount() (int, error) {
	data, err := os.ReadFile("/proc/net/tcp")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	return len(lines) - 1, nil
}

func GetUDPCount() (int, error) {
	data, err := os.ReadFile("/proc/net/udp")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	return len(lines) - 1, nil
}

func isLSB(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "LSB")
}

func GetSystemInfo() string {
	var result strings.Builder
	if isLSB("/proc/version") || isLSB("/proc/sys/kernel/ostype") {
		// for Linux
		name := getValueFromFile("/etc/os-release", "ID", "")
		result.WriteString(strings.TrimSpace(name))

		version := getValueFromFile("/etc/os-release", "VERSION_ID", "")
		if len(version) > 0 {
			version = strings.ReplaceAll(version, "\"", "")
			result.WriteString(" ")
			result.WriteString(version)
		}
	}
	return result.String()
}

func getValueFromFile(file string, key string, defaultValue string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		return defaultValue
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		values := strings.Split(strings.TrimSpace(line), "=")
		if len(values) != 2 {
			continue
		}
		if values[0] == key {
			return values[1]
		}
	}
	return defaultValue
}

func GetPid(processName string) ([]int, error) {
	// TODO:
	return nil, nil
}

func Kill(pid int) error {
	// TODO:
	return nil
}
