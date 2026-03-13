package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// getProjectRoot 获取项目根目录
func getProjectRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法获取项目根目录：%w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// getGitTags 获取所有 git tags，按版本号降序排序
func getGitTags() ([]string, error) {
	cmd := exec.Command("git", "tag", "--list", "--sort=-version:refname")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("无法获取 git tags：%w", err)
	}

	var tags []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		tag := strings.TrimSpace(scanner.Text())
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags, scanner.Err()
}
