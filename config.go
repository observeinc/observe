package main

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

var ErrCouldNotParseConfig = ObserveError{Msg: "could not parse config"}

type Config struct {
	CustomerIdStr string `json:"customerid" yaml:"customerid"`
	ClusterStr    string `json:"cluster" yaml:"cluster"`
	AuthtokenStr  string `json:"authtoken" yaml:"authtoken"`
	Quiet         bool   `json:"quiet" yaml:"quiet"`
	Debug         bool   `json:"debug" yaml:"debug"`
	Workspace     string `json:"workspace" yaml:"workspace"`
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
		return ErrCouldNotParseConfig
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
		if s.ClusterStr != "" {
			cfg.ClusterStr = s.ClusterStr
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
		return nil
	}
	if required {
		return NewObserveError(nil, "section %q not found in %q", profile, path)
	}
	return nil
}

func ReadUntypedConfig(strPath string, required bool) (map[string]any, error) {
	data, err := os.ReadFile(strPath)
	if err != nil {
		if !required {
			return nil, nil
		}
		return nil, err
	}
	buf := bytes.NewBuffer(data)
	dec := yaml.NewDecoder(buf)
	var ret map[string]any
	if err = dec.Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func SaveUntypedConfig(strPath string, config map[string]any) error {
	os.MkdirAll(path.Dir(strPath), 0775)
	buf := bytes.NewBuffer(nil)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(config); err != nil {
		return NewObserveError(err, "failed to encode config")
	}
	if err := os.WriteFile(strPath+".tmp", buf.Bytes(), 0664); err != nil {
		return NewObserveError(err, "failed to write config")
	}
	os.Remove(strPath)
	if err := os.Rename(strPath+".tmp", strPath); err != nil {
		return NewObserveError(err, "failed to save config")
	}
	return nil
}
