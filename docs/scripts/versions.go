package main

import (
	"fmt"
	"sort"
	"strings"

	semver "github.com/Masterminds/semver/v3"
)

// Version 代表一个版本文档
type Version struct {
	Name    string
	Release string
	Path    string
	Current bool
}

// createVersionsConfig 创建版本配置
func createVersionsConfig(tags []string) []Version {
	versions := make([]Version, 0, len(tags))
	for i, tag := range tags {
		version := Version{
			Name:    tag,
			Release: tag,
			Current: i == 0,
		}
		if i == 0 {
			version.Path = "/docs"
		} else {
			version.Path = fmt.Sprintf("/versioned/%s/docs", tag)
		}
		versions = append(versions, version)
	}
	return versions
}

// 按语义版本号排序（可选工具函数）
func sortVersions(tags []string) {
	sort.Slice(tags, func(i, j int) bool {
		return compareVersions(tags[i], tags[j]) > 0
	})
}

// compareVersions 比较两个版本号
// 返回 1 表示 v1 > v2，-1 表示 v1 < v2，0 表示相等
func compareVersions(v1, v2 string) int {
	ver1, ok1 := parseSemver(v1)
	ver2, ok2 := parseSemver(v2)

	switch {
	case ok1 && ok2:
		return ver1.Compare(ver2)
	case ok1:
		return 1
	case ok2:
		return -1
	default:
		return strings.Compare(strings.TrimSpace(v1), strings.TrimSpace(v2))
	}
}

func parseSemver(raw string) (*semver.Version, bool) {
	v, err := semver.StrictNewVersion(strings.TrimPrefix(strings.TrimSpace(raw), "v"))
	if err != nil {
		return nil, false
	}
	return v, true
}
