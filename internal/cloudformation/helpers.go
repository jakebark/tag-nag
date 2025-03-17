package cloudformation

import (
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// mapNodes converts a yaml mapping node into a go map
func mapNodes(node *yaml.Node) map[string]*yaml.Node {
	m := make(map[string]*yaml.Node)
	if node == nil || node.Kind != yaml.MappingNode {
		return m
	}
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		m[keyNode.Value] = valueNode
	}
	return m
}

// findMapNode parses a yaml block and returns the value, when given the key
func findMapNode(node *yaml.Node, key string) *yaml.Node {
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		k := node.Content[i]
		v := node.Content[i+1]
		if k.Value == key {
			return v
		}
	}
	return nil
}

// parseYAML unmarshal yaml and return a pointer to the root of the node
func parseYAML(filePath string) (*yaml.Node, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = *root.Content[0]
	}
	return &root, nil
}

func skipResource(node *yaml.Node, lines []string) bool {
	index := node.Line - 2
	if index < len(lines) {
		if strings.Contains(lines[index], "#tag:nag ignore") {
			return true
		}
	}
	return false
}
