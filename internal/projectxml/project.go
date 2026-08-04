package projectxml

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type Project struct {
	Path           string
	RootDir        string
	ProjectVersion string
	PackageID      string
	Version        string
	Items          []ProjectItem
	Target         TargetConfig
}

type ProjectItem struct {
	Kind    string
	Include string
}

type TargetConfig struct {
	Comparison ComparisonConfig
	Plan       PlanConfig
	Apply      ApplyConfig
}

type ComparisonConfig struct {
	IgnoredObjects []IgnoredObject
}

type IgnoredObject struct {
	Type string `xml:"Type,attr"`
	Name string `xml:"Name,attr"`
}

type PlanConfig struct {
	AllowCreate bool
	AllowAlter  bool
	AllowDrop   bool
}

type ApplyConfig struct {
	UseTransaction bool
	StopOnDataLoss bool
}

type xmlProject struct {
	XMLName        xml.Name           `xml:"D1Pac"`
	ProjectVersion string             `xml:"ProjectVersion,attr"`
	PropertyGroups []xmlPropertyGroup `xml:"PropertyGroup"`
	ItemGroups     []xmlItemGroup     `xml:"ItemGroup"`
	Target         xmlTarget          `xml:"Target"`
}

type xmlPropertyGroup struct {
	PackageID string `xml:"PackageId"`
	Version   string `xml:"Version"`
}

type xmlIncludeItem struct {
	Include string `xml:"Include,attr"`
}

type xmlItemGroup struct {
	Tables   []xmlIncludeItem `xml:"Table"`
	Indexes  []xmlIncludeItem `xml:"Index"`
	Views    []xmlIncludeItem `xml:"View"`
	Triggers []xmlIncludeItem `xml:"Trigger"`
}

type xmlTarget struct {
	Comparison xmlComparison `xml:"Comparison"`
	Plan       xmlPlan       `xml:"Plan"`
	Apply      xmlApply      `xml:"Apply"`
}

type xmlComparison struct {
	IgnoredObjects []IgnoredObject `xml:"Ignore"`
}

type xmlPlan struct {
	AllowCreate bool `xml:"AllowCreate,attr"`
	AllowAlter  bool `xml:"AllowAlter,attr"`
	AllowDrop   bool `xml:"AllowDrop,attr"`
}

type xmlApply struct {
	UseTransaction bool `xml:"UseTransaction,attr"`
	StopOnDataLoss bool `xml:"StopOnDataLossRisk,attr"`
}

type ResolvedFile struct {
	Kind    string `json:"kind"`
	AbsPath string `json:"absPath"`
	RelPath string `json:"relPath"`
}

func Load(path string) (*Project, []byte, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, err
	}
	rawXML, err := os.ReadFile(absPath)
	if err != nil {
		return nil, nil, err
	}
	return load(rawXML, absPath)
}

func LoadFromBytes(rawXML []byte, virtualPath string) (*Project, []byte, error) {
	return load(rawXML, virtualPath)
}

func load(rawXML []byte, path string) (*Project, []byte, error) {
	var document xmlProject
	if err := xml.Unmarshal(rawXML, &document); err != nil {
		return nil, nil, fmt.Errorf("invalid project xml: %w", err)
	}
	project := &Project{
		Path:           path,
		RootDir:        filepath.Dir(path),
		ProjectVersion: strings.TrimSpace(document.ProjectVersion),
		Target: TargetConfig{
			Comparison: ComparisonConfig{
				IgnoredObjects: document.Target.Comparison.IgnoredObjects,
			},
			Plan: PlanConfig{
				AllowCreate: document.Target.Plan.AllowCreate,
				AllowAlter:  document.Target.Plan.AllowAlter,
				AllowDrop:   document.Target.Plan.AllowDrop,
			},
			Apply: ApplyConfig{
				UseTransaction: document.Target.Apply.UseTransaction,
				StopOnDataLoss: document.Target.Apply.StopOnDataLoss,
			},
		},
	}
	for _, group := range document.PropertyGroups {
		if strings.TrimSpace(group.PackageID) != "" {
			project.PackageID = strings.TrimSpace(group.PackageID)
		}
		if strings.TrimSpace(group.Version) != "" {
			project.Version = strings.TrimSpace(group.Version)
		}
	}
	for _, group := range document.ItemGroups {
		project.Items = appendItems(project.Items, "table", group.Tables)
		project.Items = appendItems(project.Items, "index", group.Indexes)
		project.Items = appendItems(project.Items, "view", group.Views)
		project.Items = appendItems(project.Items, "trigger", group.Triggers)
	}
	if err := project.Validate(); err != nil {
		return nil, nil, err
	}
	return project, rawXML, nil
}

func (project *Project) Validate() error {
	if project.ProjectVersion == "" {
		return fmt.Errorf("D1Pac ProjectVersion is required")
	}
	if project.PackageID == "" {
		return fmt.Errorf("PackageId is required")
	}
	if project.Version == "" {
		return fmt.Errorf("Version is required")
	}
	if len(project.Items) == 0 {
		return fmt.Errorf("at least one ItemGroup include is required")
	}
	for _, ignored := range project.Target.Comparison.IgnoredObjects {
		switch strings.ToLower(ignored.Type) {
		case "table", "index", "view", "trigger", "*":
		default:
			return fmt.Errorf("unsupported ignored object type %q", ignored.Type)
		}
		if strings.TrimSpace(ignored.Name) == "" {
			return fmt.Errorf("ignored object Name is required")
		}
	}
	return nil
}

func (project *Project) ResolveFiles() ([]ResolvedFile, error) {
	filesystem := os.DirFS(project.RootDir)
	seen := map[string]bool{}
	var files []ResolvedFile
	for _, item := range project.Items {
		matches, err := doublestar.Glob(filesystem, item.Include)
		if err != nil {
			return nil, fmt.Errorf("invalid include %q: %w", item.Include, err)
		}
		for _, match := range matches {
			if seen[match] {
				continue
			}
			info, err := os.Stat(filepath.Join(project.RootDir, match))
			if err != nil {
				return nil, err
			}
			if info.IsDir() || !strings.EqualFold(filepath.Ext(match), ".sql") {
				continue
			}
			seen[match] = true
			files = append(files, ResolvedFile{
				Kind:    item.Kind,
				AbsPath: filepath.Join(project.RootDir, match),
				RelPath: filepath.ToSlash(match),
			})
		}
	}
	sort.Slice(files, func(left, right int) bool {
		leftWeight := kindWeight(files[left].Kind)
		rightWeight := kindWeight(files[right].Kind)
		if leftWeight == rightWeight {
			return files[left].RelPath < files[right].RelPath
		}
		return leftWeight < rightWeight
	})
	return files, nil
}

func (project *Project) IsIgnored(objectType, name string) bool {
	if strings.HasPrefix(name, "sqlite_") || strings.HasPrefix(name, "_cf_") || name == "d1_migrations" {
		return true
	}
	for _, ignored := range project.Target.Comparison.IgnoredObjects {
		if (ignored.Type == "*" || strings.EqualFold(ignored.Type, objectType)) && ignored.Name == name {
			return true
		}
	}
	return false
}

func appendItems(items []ProjectItem, kind string, includes []xmlIncludeItem) []ProjectItem {
	for _, include := range includes {
		items = append(items, ProjectItem{
			Kind:    kind,
			Include: strings.TrimSpace(include.Include),
		})
	}
	return items
}

func kindWeight(kind string) int {
	switch kind {
	case "table":
		return 10
	case "index":
		return 20
	case "view":
		return 30
	case "trigger":
		return 40
	default:
		return 100
	}
}
