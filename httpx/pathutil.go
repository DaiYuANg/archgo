package httpx

import (
	"strings"

	"github.com/samber/lo"
)

// joinRoutePath combines a normalized base path and a route fragment.
func joinRoutePath(basePath, path string) string {
	base := normalizeRoutePrefix(basePath)

	if path == "" {
		if base == "" {
			return "/"
		}
		return base
	}

	cleanPath := path
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	if base == "" {
		return cleanPath
	}

	if cleanPath == "/" {
		return base
	}

	return base + cleanPath
}

// normalizeRoutePrefix normalizes route prefixes to `\"/prefix\"` or empty.
func normalizeRoutePrefix(prefix string) string {
	clean := strings.Trim(strings.TrimSpace(prefix), "/")
	return lo.Ternary(clean == "", "", "/"+clean)
}
