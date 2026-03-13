package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// createVersionedDirs 创建版本文档目录
func createVersionedDirs(contentDir string, versions []Version) error {
	versionedDir := filepath.Join(contentDir, "versioned")
	sourceDocsDir := filepath.Join(contentDir, "docs")
	sourceRootFiles := rootFiles(contentDir)

	for _, version := range versions {
		if version.Current {
			continue
		}
		if err := createVersionDir(versionedDir, sourceDocsDir, sourceRootFiles, version); err != nil {
			return err
		}
	}

	return nil
}

func rootFiles(contentDir string) []string {
	return []string{
		filepath.Join(contentDir, "_index.md"),
		filepath.Join(contentDir, "_index.en.md"),
		filepath.Join(contentDir, "_index.zh.md"),
	}
}

func createVersionDir(versionedDir, sourceDocsDir string, sourceRootFiles []string, version Version) error {
	versionDir := filepath.Join(versionedDir, version.Name)
	versionDocsDir := filepath.Join(versionDir, "docs")

	if _, err := os.Stat(versionDir); err == nil {
		fmt.Printf("   ⏭️  跳过已存在的版本目录：%s\n", version.Name)
		return nil
	}

	fmt.Printf("   📁 创建版本文档目录：%s\n", version.Name)
	if err := os.MkdirAll(versionDocsDir, 0o755); err != nil {
		return fmt.Errorf("无法创建目录 %s：%w", versionDir, err)
	}

	if err := copyDir(sourceDocsDir, versionDocsDir); err != nil {
		fmt.Printf("      ⚠️  复制 docs 目录失败：%v\n", err)
	}

	copyRootFiles(versionDir, sourceRootFiles)
	return nil
}

func copyRootFiles(versionDir string, sourceRootFiles []string) {
	for _, srcFile := range sourceRootFiles {
		if _, err := os.Stat(srcFile); os.IsNotExist(err) {
			continue
		}
		dstFile := filepath.Join(versionDir, filepath.Base(srcFile))
		if err := copyFile(srcFile, dstFile); err != nil {
			fmt.Printf("      ⚠️  复制文件 %s 失败：%v\n", filepath.Base(srcFile), err)
		}
	}
}

// copyDir 递归复制目录
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("读取目录 %s 失败：%w", src, err)
	}

	for _, entry := range entries {
		if entry.Name() == "versioned" {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return fmt.Errorf("创建目录 %s 失败：%w", dstPath, err)
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("读取文件 %s 失败：%w", src, err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return fmt.Errorf("写入文件 %s 失败：%w", dst, err)
	}
	return nil
}
