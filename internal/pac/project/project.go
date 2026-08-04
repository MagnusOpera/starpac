package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type Item struct {
	Kind    string
	Include string
}

type ResolvedFile struct {
	Kind    string `json:"kind"`
	AbsPath string `json:"absPath"`
	RelPath string `json:"relPath"`
}

type PlanConfig struct {
	AllowCreate bool
	AllowAlter  bool
	AllowDrop   bool
}

func ValidateBase(rootName, projectVersion, packageID, version string, items []Item) error {
	if projectVersion == "" {
		return fmt.Errorf("%s ProjectVersion is required", rootName)
	}
	if packageID == "" {
		return fmt.Errorf("PackageId is required")
	}
	if version == "" {
		return fmt.Errorf("Version is required")
	}
	if len(items) == 0 {
		return fmt.Errorf("at least one ItemGroup include is required")
	}
	return nil
}

func ResolveFiles(rootDirectory string, items []Item, kindWeight func(string) int) ([]ResolvedFile, error) {
	filesystem := os.DirFS(rootDirectory)
	seen := map[string]bool{}
	var files []ResolvedFile
	for _, item := range items {
		matches, err := doublestar.Glob(filesystem, item.Include)
		if err != nil {
			return nil, fmt.Errorf("invalid include %q: %w", item.Include, err)
		}
		for _, match := range matches {
			if seen[match] {
				continue
			}
			info, err := os.Stat(filepath.Join(rootDirectory, match))
			if err != nil {
				return nil, err
			}
			if info.IsDir() || !strings.EqualFold(filepath.Ext(match), ".sql") {
				continue
			}
			seen[match] = true
			files = append(files, ResolvedFile{
				Kind:    item.Kind,
				AbsPath: filepath.Join(rootDirectory, match),
				RelPath: filepath.ToSlash(match),
			})
		}
	}
	sort.Slice(files, func(left, right int) bool {
		if kindWeight != nil {
			leftWeight := kindWeight(files[left].Kind)
			rightWeight := kindWeight(files[right].Kind)
			if leftWeight != rightWeight {
				return leftWeight < rightWeight
			}
		}
		return files[left].RelPath < files[right].RelPath
	})
	return files, nil
}
