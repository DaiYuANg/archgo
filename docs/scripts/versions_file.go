package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeVersionsFile 写入 versions.yaml 文件
func writeVersionsFile(filename string, versions []Version) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("无法创建目录：%w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("无法创建文件：%w", err)
	}
	defer file.Close()

	header := `# 版本文档配置
# 此文件定义了文档的版本列表
# 版本按时间倒序排列，第一个为当前版本

versions:
`
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("写入文件头失败：%w", err)
	}

	for i, v := range versions {
		section := "\n  # 历史版本\n"
		if v.Current {
			section = "  # 当前版本（最新版本）\n"
		}
		if _, err := file.WriteString(section); err != nil {
			return fmt.Errorf("写入章节失败：%w", err)
		}

		lines := []string{
			fmt.Sprintf("  - name: \"%s\"\n", v.Name),
			fmt.Sprintf("    release: \"%s\"\n", v.Release),
			fmt.Sprintf("    path: \"%s\"\n", v.Path),
			fmt.Sprintf("    current: %t\n", v.Current),
		}
		for _, line := range lines {
			if _, err := file.WriteString(line); err != nil {
				return fmt.Errorf("写入行失败：%w", err)
			}
		}

		if i < len(versions)-1 {
			if _, err := file.WriteString("\n"); err != nil {
				return fmt.Errorf("写入空行失败：%w", err)
			}
		}
	}

	return nil
}
