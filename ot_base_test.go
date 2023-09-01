package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInput(t *testing.T) {
	type inputFile struct {
		name    string
		content string
	}
	testcases := []struct {
		description string
		// Either filename or `-` to grab the content from stdin.
		input string
		// Optionally create the file for the test to read.
		inputFile *inputFile
		expected  *InputObject
		err       error
	}{
		{
			description: "Error on no input",
			err:         fmt.Errorf("No --definition provided"),
		},
		{
			description: "Fail to parse unsupported format",
			input:       "file.txt",
			err:         fmt.Errorf("Failed to parse input --definition"),
		},
		{
			description: "Fail to parse nonexisting yaml file",
			input:       "file.yaml",
			err:         fmt.Errorf("Provided file does not exist at path:file.yaml"),
		},
		{
			description: "Fail to parse nonexisting json file",
			input:       "file.json",
			err:         fmt.Errorf("Provided file does not exist at path:file.json"),
		},
		{
			description: "Fail to parse yaml or json input",
			input:       "random stuff",
			err:         fmt.Errorf("Failed to parse input --definition"),
		},
		{
			description: "Parse yaml input",
			input:       "-",
			inputFile: &inputFile{
				name: "/dev/stdin",
				content: `object:
  config:
    name: Default Count Check
`,
			},
			expected: &InputObject{
				Object: map[string]interface{}{
					"config": map[string]interface{}{
						"name": "Default Count Check",
					},
				},
			},
		},
		{
			description: "Parse json input",
			input:       "-",
			inputFile: &inputFile{
				name:    "/dev/stdin",
				content: `{"object":{"config": {"name": "Default Count Check"}}}`,
			},
			expected: &InputObject{
				Object: map[string]interface{}{
					"config": map[string]interface{}{
						"name": "Default Count Check",
					},
				},
			},
		},
		{
			description: "Parse yaml file",
			inputFile: &inputFile{
				name: "ot_base_temp.yaml",
				content: `object:
  config:
    name: Default Count Check
`,
			},
			expected: &InputObject{
				Object: map[string]interface{}{
					"config": map[string]interface{}{
						"name": "Default Count Check",
					},
				},
			},
		},
		{
			description: "Parse json file",
			inputFile: &inputFile{
				name:    "ot_base_temp.json",
				content: `{"object":{"config": {"name": "Default Count Check"}}}`,
			},
			expected: &InputObject{
				Object: map[string]interface{}{
					"config": map[string]interface{}{
						"name": "Default Count Check",
					},
				},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			fix := startFixture(t)
			input := tc.input
			if tc.inputFile != nil {
				if err := fix.fs.WriteFile(tc.inputFile.name, []byte(tc.inputFile.content), 0); err != nil {
					t.Error(err)
				}
				defer fix.fs.Remove(tc.inputFile.name)
				if tc.input != "-" {
					input = tc.inputFile.name
				}
			}
			obj, err := parseInput(fix.fs, fix.op, input)
			if tc.err == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, obj)
			} else {
				assert.Equal(t, tc.err, err)
			}
		})
	}
}
