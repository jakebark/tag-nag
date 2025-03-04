package cloudformation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProcessDirectory identifies all cfn files in the directory
func ProcessDirectory(dirPath string, requiredTags []string, caseInsensitive bool) int {
	var totalViolations int

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".json")) {
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

func processFile(filePath string, requiredTags []string, caseInsensitive bool) ([]Violation, error) {
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
			fmt.Printf("  %d: %s \"%s\", 🏷️ Missing tags: %v\n", v.Line, v.ResourceType, v.ResourceName, v.MissingTags)
		}
	}

	return violations, nil
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
