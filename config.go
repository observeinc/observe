package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strconv"

	"gopkg.in/yaml.v3"
)

var ErrCouldNotParseConfig = ObserveError{Msg: "could not parse config"}

type Config struct {
	CustomerIdStr     string `json:"customerid" yaml:"customerid"`
	SiteStr           string `json:"site" yaml:"site"`
	AuthtokenStr      string `json:"authtoken" yaml:"authtoken"`
	Quiet             bool   `json:"quiet" yaml:"quiet"`
	Debug             bool   `json:"debug" yaml:"debug"`
	WorkspaceIdOrName string `json:"workspace" yaml:"workspace"`
	// Don't forget to add new fields into ParseConfig(), they're not
	// automatically read into this struct!
}

func (c Config) AuthHeader() string {
	return fmt.Sprintf("Bearer %s %s", c.CustomerIdStr, c.AuthtokenStr)
}

type ConfigYaml struct {
	Profile map[string]Config `json:"profile" yaml:"profile"`
}

func ReadConfig(cfg *Config, path string, profile string, required bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if required {
			return err
		}
		return nil
	}
	return ParseConfig(data, cfg, path, profile, required)
}

func ParseConfig(data []byte, cfg *Config, path string, profile string, required bool) error {
	if len(data) == 0 && !required {
		return nil
	}
	buf := bytes.NewBuffer(data)
	dec := yaml.NewDecoder(buf)
	dec.KnownFields(true)
	var cy ConfigYaml
	if err := dec.Decode(&cy); err != nil {
		return ErrCouldNotParseConfig.WithInner(err)
	}
	if cy.Profile == nil {
		if required {
			return ErrCouldNotParseConfig
		}
		return nil
	}
	if s, has := cy.Profile[profile]; has {
		if s.CustomerIdStr != "" {
			cfg.CustomerIdStr = s.CustomerIdStr
		}
		if s.SiteStr != "" {
			cfg.SiteStr = s.SiteStr
		}
		if s.AuthtokenStr != "" {
			cfg.AuthtokenStr = s.AuthtokenStr
		}
		if s.Quiet {
			cfg.Quiet = s.Quiet
		}
		if s.Debug {
			cfg.Debug = s.Debug
		}
		if s.WorkspaceIdOrName != "" {
			cfg.WorkspaceIdOrName = s.WorkspaceIdOrName
		}
		return nil
	}
	if required {
		return NewObserveError(nil, "section %q not found in %q", profile, path)
	}
	return nil
}

func ReadUntypedConfigFromFile(fs fileSystem, strPath string, required bool) (map[string]any, error) {
	data, err := fs.ReadFile(strPath)
	if err != nil {
		if !required {
			return nil, nil
		}
		return nil, err
	}
	buf := bytes.NewBuffer(data)
	dec := yaml.NewDecoder(buf)
	var ret map[string]any
	if err := dec.Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func SaveUntypedConfig(fs fileSystem, strPath string, config map[string]any) error {
	fs.MkdirAll(path.Dir(strPath), 0775)
	buf := bytes.NewBuffer(nil)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(config); err != nil {
		return NewObserveError(err, "failed to encode config")
	}
	if err := fs.WriteFile(strPath+".tmp", buf.Bytes(), 0664); err != nil {
		return NewObserveError(err, "failed to write config")
	}
	fs.Remove(strPath)
	if err := fs.Rename(strPath+".tmp", strPath); err != nil {
		return NewObserveError(err, "failed to save config")
	}
	return nil
}

// mustGetWorkspaceId will attempt to fetch the workspace.id from the config, else it will
// query the workspaces in the customer and attempt to find it by name, else it will default
// to the first workspace in the list.
func mustGetWorkspaceId(cfg *Config, hc httpClient) string {
	// Maybe the ID is already configured.
	if cfg.WorkspaceIdOrName != "" {
		_, err := strconv.Atoi(cfg.WorkspaceIdOrName)
		isId := err == nil
		if isId {
			return cfg.WorkspaceIdOrName
		}
	}
	// WorkspaceIdOrName is a Name or empty.
	// Get all the workspaces, attempt to find one with the name=`cfg.WorkspaceIdOrName`
	var op DefaultOutput
	workspaces, err := ObjectTypeWorkspace.List(cfg, op, hc)
	if err != nil {
		panic(fmt.Sprintf("Failed to load workspaces for customer:%s", cfg.CustomerIdStr))
	}
	for _, workspace := range workspaces {
		if workspace.Name == cfg.WorkspaceIdOrName {
			return workspace.Id
		}
	}
	return workspaces[0].Id
}

// mustGetWorkspaceName will attempt to fetch the workspace.name from the config, else it will
// query the workspaces in the customer and attempt to find it by id, else it will default
// to the first workspace in the list.
func mustGetWorkspaceName(cfg *Config, hc httpClient) string {
	// Maybe the Name is already configured.
	if cfg.WorkspaceIdOrName != "" {
		_, err := strconv.Atoi(cfg.WorkspaceIdOrName)
		isName := err != nil
		if isName {
			return cfg.WorkspaceIdOrName
		}
	}
	// WorkspaceIdOrName is an ID or empty.
	// Get all the workspaces, attempt to find one with the id=`cfg.WorkspaceIdOrName`
	var op DefaultOutput
	workspaces, err := ObjectTypeWorkspace.List(cfg, op, hc)
	if err != nil {
		panic(fmt.Sprintf("Failed to load workspaces for customer:%s", cfg.CustomerIdStr))
	}
	for _, workspace := range workspaces {
		if workspace.Id == cfg.WorkspaceIdOrName {
			return workspace.Name
		}
	}
	return workspaces[0].Name
}
