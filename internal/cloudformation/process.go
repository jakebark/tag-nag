package cloudformation

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Violation struct {
	ResourceName string
	ResourceType string
	Line         int
	MissingTags  []string
}

func ProcessDirectory(dirPath string, requiredTags []string, caseInsensitive bool) int {
	totalViolations := 0

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isCloudFormationFile(path) {
			violations, err := processFile(path, requiredTags, caseInsensitive)
			if err != nil {
				fmt.Printf("Error processing file %s: %v\n", path, err)
			}
			totalViolations += len(violations)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error scanning directory:", err)
	}
	return totalViolations
}

func isCloudFormationFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml" || ext == ".json"
}

func processFile(filePath string, requiredTags []string, caseInsensitive bool) ([]Violation, error) {
	data, err := ioutil.ReadFile(filePath)
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

	var violations []Violation

	resourcesNode := findMapNode(&root, "Resources")
	if resourcesNode == nil {
		return violations, nil
	}

	for i := 0; i < len(resourcesNode.Content); i += 2 {
		resourceNameNode := resourcesNode.Content[i]
		resourceValueNode := resourcesNode.Content[i+1]
		resourceName := resourceNameNode.Value

		typeNode := findMapNode(resourceValueNode, "Type")
		if typeNode == nil {
			continue
		}
		resourceType := typeNode.Value

		if !strings.HasPrefix(resourceType, "AWS::") {
			continue
		}

		propertiesNode := findMapNode(resourceValueNode, "Properties")
		var properties map[string]interface{}
		if propertiesNode != nil {
			if err := propertiesNode.Decode(&properties); err != nil {
				properties = make(map[string]interface{})
			}
		} else {
			properties = make(map[string]interface{})
		}

		tags, err := extractTagsFromProperties(properties, caseInsensitive)
		if err != nil {
			fmt.Printf("Error extracting tags from resource %s: %v\n", resourceName, err)
			continue
		}

		// Check for missing required tags.
		missing := filterMissingTags(requiredTags, tags, caseInsensitive)
		if len(missing) > 0 {
			violations = append(violations, Violation{
				ResourceName: resourceName,
				ResourceType: resourceType,
				Line:         resourceNameNode.Line,
				MissingTags:  missing,
			})
		}
	}

	if len(violations) > 0 {
		fmt.Printf("\nViolation(s) in %s\n", filePath)
		for _, v := range violations {
			fmt.Printf("  %s \"%s\" (line %d), 🏷️ Missing tags: %v\n", v.ResourceType, v.ResourceName, v.Line, v.MissingTags)
		}
	}

	return violations, nil
}

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

func extractTagsFromProperties(properties map[string]interface{}, caseInsensitive bool) (map[string]string, error) {
	tagsMap := make(map[string]string)
	rawTags, exists := properties["Tags"]
	if !exists {
		return tagsMap, nil
	}

	tagsList, ok := rawTags.([]interface{})
	if !ok {
		return tagsMap, fmt.Errorf("Tags format is invalid")
	}

	for _, tagInterface := range tagsList {
		tagEntry, ok := tagInterface.(map[string]interface{})
		if !ok {
			continue
		}
		key, ok := tagEntry["Key"].(string)
		if !ok {
			continue
		}
		var value string
		if valStr, ok := tagEntry["Value"].(string); ok {
			value = valStr
		} else {
			if refMap, ok := tagEntry["Value"].(map[string]interface{}); ok {
				if ref, exists := refMap["Ref"]; exists {
					if refStr, ok := ref.(string); ok {
						value = fmt.Sprintf("!Ref %s", refStr)
					}
				}
			}
		}
		if caseInsensitive {
			key = strings.ToLower(key)
		}
		tagsMap[key] = value
	}
	return tagsMap, nil
}

func filterMissingTags(required []string, resourceTags map[string]string, caseInsensitive bool) []string {
	var missing []string
	for _, req := range required {
		found := false
		for tagKey := range resourceTags {
			if caseInsensitive {
				if strings.EqualFold(tagKey, req) {
					found = true
					break
				}
			} else {
				if tagKey == req {
					found = true
					break
				}
			}
		}
		if !found {
			missing = append(missing, req)
		}
	}
	return missing
}

