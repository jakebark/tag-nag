package cloudformation

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// helper to create a yaml.Node for testing
func createYamlNode(t *testing.T, yamlStr string) *yaml.Node {
	t.Helper()
	var node yaml.Node
	err := yaml.Unmarshal([]byte(yamlStr), &node)
	if err != nil {
		t.Fatalf("Failed to unmarshal test YAML: %v\nYAML:\n%s", err, yamlStr)
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return node.Content[0]
	}
	return &node
}

func testMapNodes(t *testing.T) {
	tests := []struct {
		name      string
		inputYAML string
		wantKeys  []string
	}{
		{
			name:      "simple map",
			inputYAML: `Key1: Value1\nKey2: 123`,
			wantKeys:  []string{"Key1", "Key2"},
		},
		{
			name:      "nested map",
			inputYAML: `Key1: Value1\nKey2:\n  Nested1: NestedValue`,
			wantKeys:  []string{"Key1", "Key2"},
		},
		{
			name:      "empty map",
			inputYAML: `{}`,
			wantKeys:  []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			node := createYamlNode(t, tc.inputYAML)
			mapped := mapNodes(node)
			gotKeys := make([]string, 0, len(mapped))
			for k := range mapped {
				gotKeys = append(gotKeys, k)
			}
			if len(gotKeys) != len(tc.wantKeys) {
				t.Errorf("mapNodes() returned map with %d keys, want %d. Got keys: %v", len(gotKeys), len(tc.wantKeys), gotKeys)
				return
			}
		})
	}
}

