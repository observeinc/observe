package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testConfig = `profile:
  default:
    customerid: 101
    site: observe-eng.com
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

func TestConfigYamlWorkspaceIdOrName(t *testing.T) {
	resultObject := `{ "currentUser": { "workspaces": [{ "id": "1", "name": "name-1" }, { "id": "2", "name": "name-2" }] } }`

	testcases := []struct {
		description       string
		workspaceIdOrName string
		expectedId        string
		expectedName      string
	}{
		{
			description:  "no configured workspace, default to first",
			expectedId:   "1",
			expectedName: "name-1",
		},
		{
			description:       "configured 2nd workspace incorrectly by name, default to first for id",
			workspaceIdOrName: "bleh",
			expectedId:        "1",
			expectedName:      "bleh",
		},
		{
			description:       "configured 2nd workspace incorrectly by id, default to first for name",
			workspaceIdOrName: "0",
			expectedId:        "0",
			expectedName:      "name-1",
		},
		{
			description:       "configured 2nd workspace by id",
			workspaceIdOrName: "2",
			expectedId:        "2",
			expectedName:      "name-2",
		},
		{
			description:       "configured 2nd workspace by name",
			workspaceIdOrName: "name-2",
			expectedId:        "2",
			expectedName:      "name-2",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			fix := startFixture(t,
				testRequest{"/v1/meta", 200, fmt.Sprintf(`{"data":%s}`, resultObject)},
				testRequest{"/v1/meta", 200, fmt.Sprintf(`{"data":%s}`, resultObject)},
			)

			fix.cfg.WorkspaceIdOrName = tc.workspaceIdOrName

			id := mustGetWorkspaceId(fix.cfg, fix.hc)
			assert.Equal(t, tc.expectedId, id)

			name := mustGetWorkspaceName(fix.cfg, fix.hc)
			assert.Equal(t, tc.expectedName, name)
		})
	}
}
