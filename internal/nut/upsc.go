package nut

import (
	"os/exec"
	"strings"
)

func Query(upsName string) (map[string]string, error) {
	out, err := exec.Command("upsc", upsName+"@localhost").Output()
	if err != nil {
		return nil, err
	}
	return parseUpscOutput(string(out)), nil
}

func QueryVar(upsName, varName string) (string, error) {
	out, err := exec.Command("upsc", upsName+"@localhost", varName).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func parseUpscOutput(output string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			continue
		}
		result[key] = value
	}
	return result
}
