package whisper

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const DefaultModel = "small"

type Model struct {
	Name      string
	FileName  string
	URL       string
	SHA256    string
	SHA256URL string
}

type ResolvedModel struct {
	Name          string
	Path          string
	URL           string
	SHA256        string
	SHA256URL     string
	NeedsDownload bool
	IsCustomPath  bool
}

var registry = map[string]Model{
	"tiny": {
		Name:     "tiny",
		FileName: "ggml-tiny.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin",
		SHA256:   "be07e048e1e599ad46341c8d2a135645097a538221678b7acdd1b1919c6e1b21",
	},
	"base": {
		Name:     "base",
		FileName: "ggml-base.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin",
		SHA256:   "60ed5bc3dd14eea856493d334349b405782ddcaf0028d4b5df4088345fba2efe",
	},
	"small": {
		Name:     "small",
		FileName: "ggml-small.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin",
		SHA256:   "1be3a9b2063867b937e64e2ec7483364a79917e157fa98c5d94b5c1fffea987b",
	},
	"medium": {
		Name:     "medium",
		FileName: "ggml-medium.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-medium.bin",
		SHA256:   "6c14d5adee5f86394037b4e4e8b59f1673b6cee10e3cf0b11bbdbee79c156208",
	},
	"large-v3": {
		Name:     "large-v3",
		FileName: "ggml-large-v3.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin",
		SHA256:   "64d182b440b98d5203c4f9bd541544d84c605196c4f7b845dfa11fb23594d1e2",
	},
}

func ModelNames() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func LookupModel(name string) (Model, bool) {
	model, ok := registry[name]
	return model, ok
}

func ResolveModel(modelRef, modelDir string) (ResolvedModel, error) {
	if strings.TrimSpace(modelRef) == "" {
		modelRef = DefaultModel
	}

	if model, ok := LookupModel(modelRef); ok {
		if strings.TrimSpace(modelDir) == "" {
			return ResolvedModel{}, errors.New("model directory must not be empty for named model")
		}

		modelPath := filepath.Join(modelDir, model.FileName)
		_, statErr := os.Stat(modelPath)
		needsDownload := errors.Is(statErr, os.ErrNotExist)
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return ResolvedModel{}, fmt.Errorf("stat model path: %w", statErr)
		}

		return ResolvedModel{
			Name:          model.Name,
			Path:          modelPath,
			URL:           model.URL,
			SHA256:        model.SHA256,
			SHA256URL:     model.SHA256URL,
			NeedsDownload: needsDownload,
		}, nil
	}

	if !looksLikePath(modelRef) {
		return ResolvedModel{}, fmt.Errorf("unknown model %q (known models: %s)", modelRef, strings.Join(ModelNames(), ", "))
	}

	customPath := filepath.Clean(modelRef)
	if _, err := os.Stat(customPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ResolvedModel{}, fmt.Errorf("custom model path does not exist: %s", customPath)
		}
		return ResolvedModel{}, fmt.Errorf("stat custom model path: %w", err)
	}

	return ResolvedModel{
		Path:         customPath,
		IsCustomPath: true,
	}, nil
}

func looksLikePath(input string) bool {
	return strings.ContainsRune(input, os.PathSeparator) || strings.HasSuffix(strings.ToLower(input), ".bin")
}
