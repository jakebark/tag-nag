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
		name         string
		inputYAML    string
		expectedKeys []string
	}{
		{
			name:         "simple map",
			inputYAML:    `Key1: Value1\nKey2: 123`,
			expectedKeys: []string{"Key1", "Key2"},
		},
		{
			name:         "nested map",
			inputYAML:    `Key1: Value1\nKey2:\n  Nested1: NestedValue`,
			expectedKeys: []string{"Key1", "Key2"},
		},
		{
			name:         "empty map",
			inputYAML:    `{}`,
			expectedKeys: []string{},
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
			if len(gotKeys) != len(tc.expectedKeys) {
				t.Errorf("mapNodes() returned map with %d keys, expected %d. Got keys: %v", len(gotKeys), len(tc.expectedKeys), gotKeys)
				return
			}
		})
	}
}

func testFindMapNode(t *testing.T) {
	yamlContent := `
        RootKey: RootValue
        Map1:
          NestedKey1: NestedValue1
          NestedKey2: 123
        Map2: {}
        List1:
          - itemA
          - itemB
`
	rootNode := createYamlNode(t, yamlContent)

	tests := []struct {
		name          string
		node          *yaml.Node
		key           string
		expectedNil   bool
		expectedKind  yaml.Kind
		expectedValue string
	}{
		{
			name:          "root key to root value",
			node:          rootNode,
			key:           "RootKey",
			expectedNil:   false,
			expectedKind:  yaml.ScalarNode,
			expectedValue: "RootValue",
		},
		{
			name:         "map key to map values",
			node:         rootNode,
			key:          "Map1",
			expectedNil:  false,
			expectedKind: yaml.MappingNode,
		},
		{
			name:         "map key to empty map value",
			node:         rootNode,
			key:          "Map2",
			expectedNil:  false,
			expectedKind: yaml.MappingNode,
		},
		{
			name:         "list key to list values",
			node:         rootNode,
			key:          "List1",
			expectedNil:  false,
			expectedKind: yaml.SequenceNode,
		},
		{
			name:        "missing key",
			node:        rootNode,
			key:         "MissingKey",
			expectedNil: true,
		},
		{
			name:        "search other node",
			node:        findMapNode(rootNode, "List1"),
			key:         "itemA",
			expectedNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := findMapNode(tc.node, tc.key)
			isNil := got == nil
			if isNil != tc.expectedNil {
				t.Fatalf("findMapNode(key=%q) returned nil? %t, expectedNil %t", tc.key, isNil, tc.expectedNil)
			}
			if !tc.expectedNil {
				if got.Kind != tc.expectedKind {
					t.Errorf("findMapNode(key=%q) returned node kind %v, expected %v", tc.key, got.Kind, tc.expectedKind)
				}
				if tc.expectedKind == yaml.ScalarNode && got.Value != tc.expectedValue {
					t.Errorf("findMapNode(key=%q) returned scalar value %q, expected %q", tc.key, got.Value, tc.expectedValue)
				}
			}
		})
	}
}

