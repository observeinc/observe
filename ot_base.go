package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type fileSystem interface {
	Stat(path string) (fs.FileInfo, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, b []byte, perm fs.FileMode) error
	Remove(path string) error
	Rename(oldPath, newPath string) error
	MkdirAll(path string, perm fs.FileMode) error
}

type Fs struct{}

func newFs() fileSystem {
	return Fs{}
}

func (f Fs) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

func (f Fs) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f Fs) WriteFile(path string, b []byte, perm fs.FileMode) error {
	return os.WriteFile(path, b, perm)
}

func (f Fs) Remove(path string) error {
	return os.Remove(path)
}

func (f Fs) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (f Fs) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

type workspaceObject struct {
	Id          int64
	Name        string
	Description *string
	IconUrl     *string
	WorkspaceId int64
	ManagedById *int64
}

type folderObject struct {
	FolderId int64
}

type auditedObject struct {
	CreatedBy   int64
	CreatedDate string
	UpdatedBy   int64
	UpdatedDate string
}

type InputObject struct {
	Object map[string]any `yaml:"object"`
	Params map[string]any `yaml:"params"`
}

func parseInput(fs fileSystem, op Output, flagInput string) (*InputObject, error) {
	if flagInput == "" {
		return nil, fmt.Errorf("No --definition provided")
	}
	var fileContentBytes []byte
	isYamlFile := strings.HasSuffix(flagInput, ".yaml")
	isJsonFile := strings.HasSuffix(flagInput, ".json")
	if flagInput == "-" {
		b, err := fs.ReadFile("/dev/stdin")
		if err != nil {
			return nil, err
		}
		fileContentBytes = []byte(b)
	} else if isYamlFile || isJsonFile {
		op.Debug("parseInput: parse as file\n")
		if _, err := fs.Stat(flagInput); err != nil {
			return nil, fmt.Errorf("Provided file does not exist at path:%s", flagInput)
		}
		if isYamlFile {
			op.Debug("parseInput: file is yaml\n")
		} else if isJsonFile {
			op.Debug("parseInput: file is json\n")
		}
		var err error
		fileContentBytes, err = fs.ReadFile(flagInput)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("Failed to parse input --definition")
	}

	buf := bytes.NewBuffer(fileContentBytes)
	dec := yaml.NewDecoder(buf)
	var content InputObject
	if err := dec.Decode(&content); err != nil {
		return nil, err
	}
	return &content, nil
}
