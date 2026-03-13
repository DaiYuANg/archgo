// sync-versions.go
// 版本文档同步工具
// 用途：从 git tags 自动创建版本文档目录和配置文件

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const separatorLine = "========================================"

type syncPaths struct {
	docsDir      string
	contentDir   string
	versionsFile string
}

func main() {
	printBanner()

	paths := resolvePaths()
	tags := loadTags()
	latestTag := tags[0]

	printCurrentVersion(latestTag)

	versions := createVersionsConfig(tags)
	writeVersions(paths.versionsFile, versions)
	createVersionDirs(paths.contentDir, versions)

	printSummary(latestTag, len(tags))
}

func printBanner() {
	fmt.Println(separatorLine)
	fmt.Println("   ArcGo 版本文档同步工具")
	fmt.Println(separatorLine)
	fmt.Println()
}

func resolvePaths() syncPaths {
	projectRoot, err := getProjectRoot()
	if err != nil {
		exitWithError(err)
	}

	docsDir := filepath.Join(projectRoot, "docs")
	return syncPaths{
		docsDir:      docsDir,
		contentDir:   filepath.Join(docsDir, "content"),
		versionsFile: filepath.Join(docsDir, "data", "versions.yaml"),
	}
}

func loadTags() []string {
	fmt.Println("[1/4] 获取 git tags...")
	tags, err := getGitTags()
	if err != nil {
		exitWithError(err)
	}
	if len(tags) == 0 {
		exitWithMessage("❌ 没有找到任何 git tags")
	}

	fmt.Println("✅ 找到以下 tags:")
	for _, tag := range tags {
		fmt.Printf("   - %s\n", tag)
	}
	fmt.Println()
	return tags
}

func printCurrentVersion(latestTag string) {
	fmt.Printf("[2/4] 当前版本：%s\n", latestTag)
	fmt.Println()
}

func writeVersions(versionsFile string, versions []Version) {
	fmt.Println("[3/4] 创建版本配置文件...")
	if err := writeVersionsFile(versionsFile, versions); err != nil {
		exitWithError(err)
	}
	fmt.Printf("✅ 版本文档配置已更新到：%s\n", versionsFile)
	fmt.Println()
}

func createVersionDirs(contentDir string, versions []Version) {
	fmt.Println("[4/4] 创建版本文档目录...")
	if err := createVersionedDirs(contentDir, versions); err != nil {
		exitWithError(err)
	}
	fmt.Println()
}

func printSummary(latestTag string, total int) {
	fmt.Println(separatorLine)
	fmt.Println("   版本统计")
	fmt.Println(separatorLine)
	fmt.Printf("   当前版本：%s\n", latestTag)
	fmt.Printf("   历史版本数：%d 个\n", total)
	fmt.Println(separatorLine)
	fmt.Println()
	fmt.Println("💡 提示：运行 'go tool hugo server -D' 预览版本文档")
	fmt.Println()
}

func exitWithError(err error) {
	fmt.Printf("❌ 错误：%v\n", err)
	os.Exit(1)
}

func exitWithMessage(message string) {
	fmt.Println(message)
	os.Exit(1)
}
