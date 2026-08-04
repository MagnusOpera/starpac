package packagefmt

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sharedarchive "github.com/MagnusOpera/starpac/internal/pac/archive"
	"github.com/MagnusOpera/starpac/internal/postgres/model"
	"github.com/MagnusOpera/starpac/internal/postgres/project"
)

type Manifest struct {
	FormatVersion   string         `json:"formatVersion"`
	PackageID       string         `json:"packageId"`
	PackageVersion  string         `json:"packageVersion"`
	PostgresVersion int            `json:"postgresVersion"`
	BuiltAtUTC      time.Time      `json:"builtAtUtc"`
	ProjectFile     string         `json:"projectFile"`
	Files           []ManifestFile `json:"files"`
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

func NewManifest(project *projectxml.Project, schemaModel *model.SchemaModel) Manifest {
	_ = schemaModel
	return Manifest{
		FormatVersion:   "1",
		PackageID:       project.PackageID,
		PackageVersion:  project.Version,
		PostgresVersion: project.PostgresVersion,
		BuiltAtUTC:      time.Now().UTC(),
		ProjectFile:     filepath.Base(project.Path),
	}
}

func Write(path string, manifest Manifest, schemaModel *model.SchemaModel, rawProject []byte, project *projectxml.Project) error {
	files, err := project.ResolveFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		content, err := os.ReadFile(file.AbsPath)
		if err != nil {
			return err
		}
		manifest.Files = append(manifest.Files, ManifestFile{
			Kind:         file.Kind,
			RelativePath: file.RelPath,
			SHA256:       sharedarchive.Checksum(content),
		})
	}

	model.Sort(schemaModel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	if err := sharedarchive.WriteJSON(zipWriter, "manifest.json", manifest); err != nil {
		return err
	}
	if err := sharedarchive.WriteJSON(zipWriter, "model.json", schemaModel); err != nil {
		return err
	}
	if err := sharedarchive.WriteBytes(zipWriter, "project.xml", rawProject); err != nil {
		return err
	}

	var checksumLines []string
	for _, file := range files {
		content, err := os.ReadFile(file.AbsPath)
		if err != nil {
			return err
		}
		entryName := "scripts/" + filepath.ToSlash(file.RelPath)
		if err := sharedarchive.WriteBytes(zipWriter, entryName, content); err != nil {
			return err
		}
		checksumLines = append(checksumLines, sharedarchive.Checksum(content)+"  "+entryName)
	}

	return sharedarchive.WriteBytes(zipWriter, "checksums/files.sha256", []byte(strings.Join(checksumLines, "\n")+"\n"))
}

func Read(path string) (*Package, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	pkg := &Package{Scripts: map[string][]byte{}}
	var projectXML []byte
	for _, file := range reader.File {
		data, err := sharedarchive.ReadFile(file)
		if err != nil {
			return nil, err
		}
		switch file.Name {
		case "manifest.json":
			if err := json.Unmarshal(data, &pkg.Manifest); err != nil {
				return nil, err
			}
		case "model.json":
			var schemaModel model.SchemaModel
			if err := json.Unmarshal(data, &schemaModel); err != nil {
				return nil, err
			}
			pkg.Model = &schemaModel
		case "project.xml":
			projectXML = data
		default:
			if strings.HasPrefix(file.Name, "scripts/") {
				pkg.Scripts[file.Name] = data
			}
		}
	}

	if pkg.Model == nil {
		return nil, fmt.Errorf("package is missing model.json")
	}
	if projectXML == nil {
		return nil, fmt.Errorf("package is missing project.xml")
	}

	project, _, err := projectxml.LoadFromBytes(projectXML, pkg.Manifest.ProjectFile)
	if err != nil {
		return nil, err
	}
	pkg.Project = project
	pkg.ProjectXML = projectXML
	return pkg, nil
}
