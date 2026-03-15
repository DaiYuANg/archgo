package configx

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// ErrUnsupportedFileFormat is returned when a config file has an extension
// that configx does not know how to parse. Supported extensions are:
// .yaml, .yml, .json, .toml.
var ErrUnsupportedFileFormat = errors.New("configx: unsupported config file format")

// supportedExtensions lists every extension that loadFiles can parse.
var supportedExtensions = []string{".yaml", ".yml", ".json", ".toml"}

// parserFor returns the koanf.Parser for the given file extension, or nil if
// the extension is not supported. Callers must check for nil before use.
func parserFor(ext string) koanf.Parser {
	switch ext {
	case ".yaml", ".yml":
		return yaml.Parser()
	case ".json":
		return json.Parser()
	case ".toml":
		return toml.Parser()
	default:
		return nil
	}
}

// loadFiles loads each file in order into k, merging on top of any previously
// loaded values. Later files take precedence over earlier ones.
//
// Returns [ErrUnsupportedFileFormat] (wrapped) if any file has an extension
// that is not in [supportedExtensions]. Use errors.Is to detect it.
func loadFiles(k *koanf.Koanf, files []string) error {
	for _, f := range files {
		ext := filepath.Ext(f)

		parser := parserFor(ext)
		if parser == nil {
			return fmt.Errorf("%w: %q (got %q, want one of %v)",
				ErrUnsupportedFileFormat, f, ext, supportedExtensions)
		}

		if err := k.Load(file.Provider(f), parser); err != nil {
			return fmt.Errorf("configx: load config file %q: %w", f, err)
		}
	}
	return nil
}
