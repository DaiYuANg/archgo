package configx

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	envProvider "github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

// loadDotenv loads related configuration.
// ignoreErr documents related behavior.
func loadDotenv(k *koanf.Koanf, files []string, ignoreErr bool) error {
	for _, f := range files {
		// Note.
		if _, err := os.Stat(f); os.IsNotExist(err) {
			if !ignoreErr {
				return err
			}
			// Note.
			continue
		}
		if err := godotenv.Load(f); err != nil && !ignoreErr {
			return err
		}
		// Note.
	}
	return nil
}

// loadEnv loads related configuration.
// prefix documents related behavior.
// Note.
// Note.
func loadEnv(k *koanf.Koanf, prefix string) error {
	normalizedPrefix := normalizeEnvPrefix(prefix)

	p := envProvider.Provider(".", envProvider.Opt{
		Prefix: normalizedPrefix,
		TransformFunc: func(k, v string) (string, any) {
			keyWithoutPrefix := strings.TrimPrefix(k, normalizedPrefix)
			keyWithoutPrefix = strings.TrimPrefix(keyWithoutPrefix, "_")

			// Note.
			key := strings.ReplaceAll(strings.ToLower(keyWithoutPrefix), "_", ".")
			return key, v
		},
		EnvironFunc: os.Environ,
	})

	return k.Load(p, nil)
}

func normalizeEnvPrefix(prefix string) string {
	clean := strings.TrimSpace(prefix)
	if clean == "" {
		return ""
	}
	if strings.HasSuffix(clean, "_") {
		return clean
	}
	return clean + "_"
}
