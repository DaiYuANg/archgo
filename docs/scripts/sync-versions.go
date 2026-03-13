// sync-versions.go
// 版本文档同步工具
// 用途：从 git tags 自动创建版本文档目录和配置文件

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("   ArcGo 版本文档同步工具")
	fmt.Println("========================================")
	fmt.Println()

	projectRoot, err := getProjectRoot()
	if err != nil {
		fmt.Printf("❌ 错误：%v\n", err)
		os.Exit(1)
	}

	docsDir := filepath.Join(projectRoot, "docs")
	contentDir := filepath.Join(docsDir, "content")
	versionsFile := filepath.Join(docsDir, "data", "versions.yaml")

	fmt.Println("[1/4] 获取 git tags...")
	tags, err := getGitTags()
	if err != nil {
		fmt.Printf("❌ 错误：%v\n", err)
		os.Exit(1)
	}
	if len(tags) == 0 {
		fmt.Println("❌ 没有找到任何 git tags")
		os.Exit(1)
	}

	fmt.Println("✅ 找到以下 tags:")
	for _, tag := range tags {
		fmt.Printf("   - %s\n", tag)
	}
	fmt.Println()

	latestTag := tags[0]
	fmt.Printf("[2/4] 当前版本：%s\n", latestTag)
	fmt.Println()

	fmt.Println("[3/4] 创建版本配置文件...")
	versions := createVersionsConfig(tags)
	if err := writeVersionsFile(versionsFile, versions); err != nil {
		fmt.Printf("❌ 错误：%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 版本文档配置已更新到：%s\n", versionsFile)
	fmt.Println()

	fmt.Println("[4/4] 创建版本文档目录...")
	if err := createVersionedDirs(contentDir, versions); err != nil {
		fmt.Printf("❌ 错误：%v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("   版本统计")
	fmt.Println("========================================")
	fmt.Printf("   当前版本：%s\n", latestTag)
	fmt.Printf("   历史版本数：%d 个\n", len(tags))
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("💡 提示：运行 'go tool hugo server -D' 预览版本文档")
	fmt.Println()
}
