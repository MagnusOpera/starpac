package projectxml

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type Project struct {
	Path            string
	RootDir         string
	ProjectVersion  string
	PackageID       string
	Version         string
	PostgresVersion int
	DefaultSchema   string
	Items           []ProjectItem
	Target          TargetConfig
}

type ProjectItem struct {
	Kind    string
	Include string
}

type TargetConfig struct {
	OwnedSchemas []OwnedSchema
	Extensions   []TargetExtension
	Comparison   ComparisonConfig
	Plan         PlanConfig
	Apply        ApplyConfig
}

type OwnedSchema struct {
	Name string `xml:"Name,attr"`
}

type TargetExtension struct {
	Name    string `xml:"Name,attr"`
	Version string `xml:"Version,attr"`
}

type ComparisonConfig struct {
	MatchPrivileges bool
	MatchOwners     bool
	MatchComments   bool
}

type PlanConfig struct {
	AllowCreate bool
	AllowAlter  bool
	AllowDrop   bool
}

type ApplyConfig struct {
	UseTransaction   bool
	LockTimeout      string
	StatementTimeout string
	StopOnDataLoss   bool
}

type xmlProject struct {
	XMLName        xml.Name           `xml:"PgPac"`
	ProjectVersion string             `xml:"ProjectVersion,attr"`
	PropertyGroup  []xmlPropertyGroup `xml:"PropertyGroup"`
	ItemGroup      []xmlItemGroup     `xml:"ItemGroup"`
	Target         xmlTarget          `xml:"Target"`
}

type xmlPropertyGroup struct {
	PackageID       string `xml:"PackageId"`
	Version         string `xml:"Version"`
	PostgresVersion int    `xml:"PostgresVersion"`
	DefaultSchema   string `xml:"DefaultSchema"`
}

type xmlIncludeItem struct {
	Include string `xml:"Include,attr"`
}

type xmlItemGroup struct {
	Schema    []xmlIncludeItem `xml:"Schema"`
	Table     []xmlIncludeItem `xml:"Table"`
	View      []xmlIncludeItem `xml:"View"`
	Function  []xmlIncludeItem `xml:"Function"`
	Type      []xmlIncludeItem `xml:"Type"`
	Extension []xmlIncludeItem `xml:"Extension"`
	Security  []xmlIncludeItem `xml:"Security"`
}

type xmlTarget struct {
	OwnedSchemas xmlOwnedSchemas `xml:"OwnedSchemas"`
	Extensions   xmlExtensions   `xml:"Extensions"`
	Comparison   xmlComparison   `xml:"Comparison"`
	Plan         xmlPlan         `xml:"Plan"`
	Apply        xmlApply        `xml:"Apply"`
}

type xmlOwnedSchemas struct {
	Schemas []OwnedSchema `xml:"Schema"`
}

type xmlExtensions struct {
	Extensions []TargetExtension `xml:"Extension"`
}

type xmlComparison struct {
	MatchPrivileges bool `xml:"MatchPrivileges,attr"`
	MatchOwners     bool `xml:"MatchOwners,attr"`
	MatchComments   bool `xml:"MatchComments,attr"`
}

type xmlPlan struct {
	AllowCreate bool `xml:"AllowCreate,attr"`
	AllowAlter  bool `xml:"AllowAlter,attr"`
	AllowDrop   bool `xml:"AllowDrop,attr"`
}

type xmlApply struct {
	UseTransaction   bool   `xml:"UseTransaction,attr"`
	LockTimeout      string `xml:"LockTimeout,attr"`
	StatementTimeout string `xml:"StatementTimeout,attr"`
	StopOnDataLoss   bool   `xml:"StopOnDataLossRisk,attr"`
}

type ResolvedFile struct {
	Kind    string `json:"kind"`
	AbsPath string `json:"absPath"`
	RelPath string `json:"relPath"`
	Content string `json:"-"`
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

	var doc xmlProject
	if err := xml.Unmarshal(rawXML, &doc); err != nil {
		return nil, nil, fmt.Errorf("invalid project xml: %w", err)
	}

	project := &Project{
		Path:           absPath,
		RootDir:        filepath.Dir(absPath),
		ProjectVersion: strings.TrimSpace(doc.ProjectVersion),
		Target: TargetConfig{
			OwnedSchemas: doc.Target.OwnedSchemas.Schemas,
			Extensions:   doc.Target.Extensions.Extensions,
			Comparison: ComparisonConfig{
				MatchPrivileges: doc.Target.Comparison.MatchPrivileges,
				MatchOwners:     doc.Target.Comparison.MatchOwners,
				MatchComments:   doc.Target.Comparison.MatchComments,
			},
			Plan: PlanConfig{
				AllowCreate: doc.Target.Plan.AllowCreate,
				AllowAlter:  doc.Target.Plan.AllowAlter,
				AllowDrop:   doc.Target.Plan.AllowDrop,
			},
			Apply: ApplyConfig{
				UseTransaction:   doc.Target.Apply.UseTransaction,
				LockTimeout:      doc.Target.Apply.LockTimeout,
				StatementTimeout: doc.Target.Apply.StatementTimeout,
				StopOnDataLoss:   doc.Target.Apply.StopOnDataLoss,
			},
		},
	}

	for _, group := range doc.PropertyGroup {
		if group.PackageID != "" {
			project.PackageID = strings.TrimSpace(group.PackageID)
		}
		if group.Version != "" {
			project.Version = strings.TrimSpace(group.Version)
		}
		if group.PostgresVersion != 0 {
			project.PostgresVersion = group.PostgresVersion
		}
		if group.DefaultSchema != "" {
			project.DefaultSchema = strings.TrimSpace(group.DefaultSchema)
		}
	}

	for _, group := range doc.ItemGroup {
		project.Items = appendIncludeItems(project.Items, "schema", group.Schema)
		project.Items = appendIncludeItems(project.Items, "table", group.Table)
		project.Items = appendIncludeItems(project.Items, "view", group.View)
		project.Items = appendIncludeItems(project.Items, "function", group.Function)
		project.Items = appendIncludeItems(project.Items, "type", group.Type)
		project.Items = appendIncludeItems(project.Items, "extension", group.Extension)
		project.Items = appendIncludeItems(project.Items, "security", group.Security)
	}

	if err := project.Validate(); err != nil {
		return nil, nil, err
	}

	return project, rawXML, nil
}

func LoadFromBytes(rawXML []byte, virtualPath string) (*Project, []byte, error) {
	var doc xmlProject
	if err := xml.Unmarshal(rawXML, &doc); err != nil {
		return nil, nil, fmt.Errorf("invalid project xml: %w", err)
	}

	project := &Project{
		Path:           virtualPath,
		RootDir:        filepath.Dir(virtualPath),
		ProjectVersion: strings.TrimSpace(doc.ProjectVersion),
		Target: TargetConfig{
			OwnedSchemas: doc.Target.OwnedSchemas.Schemas,
			Extensions:   doc.Target.Extensions.Extensions,
			Comparison: ComparisonConfig{
				MatchPrivileges: doc.Target.Comparison.MatchPrivileges,
				MatchOwners:     doc.Target.Comparison.MatchOwners,
				MatchComments:   doc.Target.Comparison.MatchComments,
			},
			Plan: PlanConfig{
				AllowCreate: doc.Target.Plan.AllowCreate,
				AllowAlter:  doc.Target.Plan.AllowAlter,
				AllowDrop:   doc.Target.Plan.AllowDrop,
			},
			Apply: ApplyConfig{
				UseTransaction:   doc.Target.Apply.UseTransaction,
				LockTimeout:      doc.Target.Apply.LockTimeout,
				StatementTimeout: doc.Target.Apply.StatementTimeout,
				StopOnDataLoss:   doc.Target.Apply.StopOnDataLoss,
			},
		},
	}

	for _, group := range doc.PropertyGroup {
		if group.PackageID != "" {
			project.PackageID = strings.TrimSpace(group.PackageID)
		}
		if group.Version != "" {
			project.Version = strings.TrimSpace(group.Version)
		}
		if group.PostgresVersion != 0 {
			project.PostgresVersion = group.PostgresVersion
		}
		if group.DefaultSchema != "" {
			project.DefaultSchema = strings.TrimSpace(group.DefaultSchema)
		}
	}
	for _, group := range doc.ItemGroup {
		project.Items = appendIncludeItems(project.Items, "schema", group.Schema)
		project.Items = appendIncludeItems(project.Items, "table", group.Table)
		project.Items = appendIncludeItems(project.Items, "view", group.View)
		project.Items = appendIncludeItems(project.Items, "function", group.Function)
		project.Items = appendIncludeItems(project.Items, "type", group.Type)
		project.Items = appendIncludeItems(project.Items, "extension", group.Extension)
		project.Items = appendIncludeItems(project.Items, "security", group.Security)
	}
	if err := project.Validate(); err != nil {
		return nil, nil, err
	}
	return project, rawXML, nil
}

func appendIncludeItems(items []ProjectItem, kind string, xmlItems []xmlIncludeItem) []ProjectItem {
	for _, item := range xmlItems {
		items = append(items, ProjectItem{Kind: kind, Include: item.Include})
	}
	return items
}

func (p *Project) Validate() error {
	if p.ProjectVersion == "" {
		return fmt.Errorf("PgPac ProjectVersion is required")
	}
	if p.PackageID == "" {
		return fmt.Errorf("PackageId is required")
	}
	if p.Version == "" {
		return fmt.Errorf("Version is required")
	}
	if p.PostgresVersion < 17 {
		return fmt.Errorf("PostgresVersion must be 17 or newer, found %d", p.PostgresVersion)
	}
	if p.DefaultSchema == "" {
		return fmt.Errorf("DefaultSchema is required")
	}
	if len(p.Items) == 0 {
		return fmt.Errorf("at least one ItemGroup include is required")
	}
	if len(p.Target.OwnedSchemas) == 0 {
		return fmt.Errorf("Target/OwnedSchemas must contain at least one schema")
	}
	return nil
}

func (p *Project) ResolveFiles() ([]ResolvedFile, error) {
	filesystem := os.DirFS(p.RootDir)
	seen := map[string]bool{}
	var files []ResolvedFile
	for _, item := range p.Items {
		matches, err := doublestar.Glob(filesystem, item.Include)
		if err != nil {
			return nil, fmt.Errorf("invalid include %q: %w", item.Include, err)
		}
		for _, match := range matches {
			if seen[match] {
				continue
			}
			info, err := os.Stat(filepath.Join(p.RootDir, match))
			if err != nil {
				return nil, err
			}
			if info.IsDir() {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(match), ".sql") {
				continue
			}
			seen[match] = true
			files = append(files, ResolvedFile{
				Kind:    item.Kind,
				AbsPath: filepath.Join(p.RootDir, match),
				RelPath: filepath.ToSlash(match),
			})
		}
	}
	sortResolvedFiles(files)
	return files, nil
}

func sortResolvedFiles(files []ResolvedFile) {
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[j].RelPath < files[i].RelPath {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
}

func (t TargetConfig) OwnedSchemaNames() []string {
	names := make([]string, 0, len(t.OwnedSchemas))
	for _, schema := range t.OwnedSchemas {
		names = append(names, schema.Name)
	}
	return names
}

func (t TargetConfig) ExtensionNames() []string {
	names := make([]string, 0, len(t.Extensions))
	for _, extension := range t.Extensions {
		names = append(names, extension.Name)
	}
	return names
}
