package main

import "testing"

const testConfig = `profile:
  default:
    customerid: 101
    cluster: observe-eng.com
    authtoken: not-an-authtoken
`

func TestConfigYaml(t *testing.T) {
	var cfg Config
	e := ParseConfig([]byte(""), &cfg, "some path", "default", false)
	if e != nil {
		t.Errorf("Expected nil, got %v", e)
	}
	e = ParseConfig([]byte(""), &cfg, "some path", "default", true)
	if e == nil {
		t.Errorf("Expected error, got nil")
	}
	e = ParseConfig([]byte(testConfig), &cfg, "some path", "default", true)
	if e != nil {
		t.Errorf("Expected nil, got %v", e)
	}
}
