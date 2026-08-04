package packagefmt

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MagnusOpera/starpac/internal/d1/model"
	"github.com/MagnusOpera/starpac/internal/d1/project"
	sharedarchive "github.com/MagnusOpera/starpac/internal/pac/archive"
)

type Manifest struct {
	FormatVersion  string         `json:"formatVersion"`
	Engine         string         `json:"engine"`
	PackageID      string         `json:"packageId"`
	PackageVersion string         `json:"packageVersion"`
	SQLiteVersion  string         `json:"sqliteVersion"`
	BuiltAtUTC     time.Time      `json:"builtAtUtc"`
	ProjectFile    string         `json:"projectFile"`
	Files          []ManifestFile `json:"files"`
}

type ManifestFile struct {
	Kind         string `json:"kind"`
	RelativePath string `json:"relativePath"`
	SHA256       string `json:"sha256"`
}

type Package struct {
	Manifest   Manifest
	Model      *model.SchemaModel
	Project    *projectxml.Project
	ProjectXML []byte
	Scripts    map[string][]byte
}

func NewManifest(project *projectxml.Project, schema *model.SchemaModel) Manifest {
	return Manifest{
		FormatVersion:  "1",
		Engine:         "cloudflare-d1",
		PackageID:      project.PackageID,
		PackageVersion: project.Version,
		SQLiteVersion:  schema.SQLiteVersion,
		BuiltAtUTC:     time.Now().UTC(),
		ProjectFile:    filepath.Base(project.Path),
	}
}

func Write(
	path string,
	manifest Manifest,
	schema *model.SchemaModel,
	rawProject []byte,
	project *projectxml.Project,
) error {
	files, err := project.ResolveFiles()
	if err != nil {
		return err
	}
	contents := make(map[string][]byte, len(files))
	for _, file := range files {
		content, err := os.ReadFile(file.AbsPath)
		if err != nil {
			return err
		}
		contents[file.RelPath] = content
		manifest.Files = append(manifest.Files, ManifestFile{
			Kind:         file.Kind,
			RelativePath: file.RelPath,
			SHA256:       sharedarchive.Checksum(content),
		})
	}
	model.Sort(schema)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()
	archive := zip.NewWriter(output)
	defer archive.Close()
	if err := sharedarchive.WriteJSON(archive, "manifest.json", manifest); err != nil {
		return err
	}
	if err := sharedarchive.WriteJSON(archive, "model.json", schema); err != nil {
		return err
	}
	if err := sharedarchive.WriteBytes(archive, "project.xml", rawProject); err != nil {
		return err
	}
	var checksumLines []string
	for _, file := range files {
		entryName := "scripts/" + filepath.ToSlash(file.RelPath)
		content := contents[file.RelPath]
		if err := sharedarchive.WriteBytes(archive, entryName, content); err != nil {
			return err
		}
		checksumLines = append(checksumLines, sharedarchive.Checksum(content)+"  "+entryName)
	}
	return sharedarchive.WriteBytes(
		archive,
		"checksums/files.sha256",
		[]byte(strings.Join(checksumLines, "\n")+"\n"),
	)
}

func Read(path string) (*Package, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	result := &Package{Scripts: map[string][]byte{}}
	for _, file := range archive.File {
		content, err := sharedarchive.ReadFile(file)
		if err != nil {
			return nil, err
		}
		switch file.Name {
		case "manifest.json":
			if err := json.Unmarshal(content, &result.Manifest); err != nil {
				return nil, err
			}
		case "model.json":
			var schema model.SchemaModel
			if err := json.Unmarshal(content, &schema); err != nil {
				return nil, err
			}
			result.Model = &schema
		case "project.xml":
			result.ProjectXML = content
		default:
			if strings.HasPrefix(file.Name, "scripts/") {
				result.Scripts[file.Name] = content
			}
		}
	}
	if result.Manifest.FormatVersion != "1" {
		return nil, fmt.Errorf("unsupported package format version %q", result.Manifest.FormatVersion)
	}
	if result.Manifest.Engine != "cloudflare-d1" {
		return nil, fmt.Errorf("package engine must be cloudflare-d1, found %q", result.Manifest.Engine)
	}
	if result.Model == nil {
		return nil, fmt.Errorf("package is missing model.json")
	}
	if result.ProjectXML == nil {
		return nil, fmt.Errorf("package is missing project.xml")
	}
	project, _, err := projectxml.LoadFromBytes(result.ProjectXML, result.Manifest.ProjectFile)
	if err != nil {
		return nil, err
	}
	result.Project = project
	return result, nil
}
