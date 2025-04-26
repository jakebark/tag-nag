package inputs

import (
	"reflect"
	"testing"

	"github.com/jakebark/tag-nag/internal/shared"
)

func testSplitTags(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", []string{""}},
		{"key", "Owner", []string{"Owner"}},
		{"multiple keys", "Owner, Env , Project", []string{"Owner", "Env", "Project"}},
		{"value", "Owner[Jake]", []string{"Owner[Jake]"}},
		{"multiple values", "Env[Dev,Prod]", []string{"Env[Dev,Prod]"}},
		{"mixed keys and values", "Owner[Jake], Env[Dev,Prod], CostCenter", []string{"Owner[Jake]", "Env[Dev,Prod]", "CostCenter"}},
		{"legacy value input", "owner=jake, env=prod", []string{"owner=jake", "env=prod"}},
		{"mixed legacy", "owner=jake, Env[Dev,Prod]", []string{"owner=jake", "Env[Dev,Prod]"}},
		{"trailing comma", "Owner,Env,", []string{"Owner", "Env", ""}},
		{"leading comma", ",Owner,Env", []string{"", "Owner", "Env"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := splitTags(tc.input)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("splitTags(%q) = %v; want %v", tc.input, actual, tc.expected)
			}
		})
	}
}
