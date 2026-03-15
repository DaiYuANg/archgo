package configx

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	envProvider "github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

// loadDotenv loads each dotenv file in order. If ignoreErr is true, missing
// files and parse errors are silently skipped; otherwise they are returned as
// errors.
func loadDotenv(files []string, ignoreErr bool) error {
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			if ignoreErr {
				continue
			}
			if os.IsNotExist(err) {
				return fmt.Errorf("configx: dotenv file %q not found: %w", f, err)
			}
			return fmt.Errorf("configx: stat dotenv file %q: %w", f, err)
		}

		if err := godotenv.Load(f); err != nil {
			if ignoreErr {
				continue
			}
			return fmt.Errorf("configx: load dotenv file %q: %w", f, err)
		}
	}
	return nil
}

// loadEnv loads environment variables into k. Only variables whose names begin
// with the given prefix (e.g. "APP_") are considered.
//
// The separator controls how the remainder of the key is translated into a
// koanf path:
//
//   - separator "_"  (default): every underscore becomes ".", so APP_DB_HOST
//     becomes the path "db.host".
//   - separator "__" (double-underscore convention): only double underscores
//     become ".", so APP_DB__HOST → "db.host" while APP_MAX_RETRY → "max_retry".
//
// Keys and values are always lowercased before insertion.
func loadEnv(k *koanf.Koanf, prefix, separator string) error {
	if separator == "" {
		separator = defaultEnvSeparator
	}

	normalizedPrefix := normalizeEnvPrefix(prefix)

	p := envProvider.Provider(".", envProvider.Opt{
		Prefix: normalizedPrefix,
		TransformFunc: func(k, v string) (string, any) {
			keyWithoutPrefix := strings.TrimPrefix(k, normalizedPrefix)
			keyWithoutPrefix = strings.TrimPrefix(keyWithoutPrefix, "_")

			// Replace the chosen separator with "." to form a koanf path.
			// The key is lowercased so that APP_DB_HOST and app_db_host are
			// treated identically.
			key := strings.ReplaceAll(
				strings.ToLower(keyWithoutPrefix),
				strings.ToLower(separator),
				".",
			)
			return key, v
		},
		EnvironFunc: os.Environ,
	})

	if err := k.Load(p, nil); err != nil {
		return fmt.Errorf("configx: load env prefix %q: %w", normalizedPrefix, err)
	}
	return nil
}

// normalizeEnvPrefix ensures the prefix ends with exactly one trailing
// underscore. An empty prefix is returned as-is so that all env vars are
// considered.
func normalizeEnvPrefix(prefix string) string {
	clean := strings.TrimSpace(prefix)
	if clean == "" {
		return ""
	}
	return strings.TrimSuffix(clean, "_") + "_"
}
