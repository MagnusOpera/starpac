package archive

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
)

func WriteJSON(archive *zip.Writer, name string, value any) error {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return err
	}
	return WriteBytes(archive, name, buffer.Bytes())
}

func WriteBytes(archive *zip.Writer, name string, content []byte) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	writer, err := archive.Create(name)
	if err != nil {
		return err
	}
	_, err = writer.Write(content)
	return err
}

func ReadFile(file *zip.File) ([]byte, error) {
	if err := ValidateName(file.Name); err != nil {
		return nil, err
	}
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func Checksum(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func ValidateName(name string) error {
	cleaned := path.Clean(strings.ReplaceAll(name, "\\", "/"))
	if name == "" || strings.HasPrefix(cleaned, "/") || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return fmt.Errorf("unsafe package entry name %q", name)
	}
	return nil
}
