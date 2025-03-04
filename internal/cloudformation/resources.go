package cloudformation

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// getResourceViolations inspects resource blocks and returns violations
func checkResourcesforTags(resourcesMapping map[string]*yaml.Node, requiredTags []string, caseInsensitive bool) []Violation {
	var violations []Violation
	for resourceName, resourceNode := range resourcesMapping { // resourceNode == yaml node for resource
		resourceMapping := mapNodes(resourceNode)

		typeNode, ok := resourceMapping["Type"]
		if !ok || !strings.HasPrefix(typeNode.Value, "AWS::") {
			continue
		}
		resourceType := typeNode.Value

		properties := make(map[string]interface{}) // tags are part of the properties  node
		if propsNode, ok := resourceMapping["Properties"]; ok {
			_ = propsNode.Decode(&properties)
		}

		tags, err := extractTagMap(properties, caseInsensitive)
		if err != nil {
			fmt.Printf("Error extracting tags from resource %s: %v\n", resourceName, err)
			continue
		}

		missing := filterMissingTags(requiredTags, tags, caseInsensitive)
		if len(missing) > 0 {
			violations = append(violations, Violation{
				ResourceName: resourceName,
				ResourceType: resourceType,
				Line:         resourceNode.Line,
				MissingTags:  missing,
			})
		}
	}
	return violations
}

// extractTagMap extracts a yaml/json map to a go map
func extractTagMap(properties map[string]interface{}, caseInsensitive bool) (map[string]string, error) {
	tagsMap := make(map[string]string)
	literalTags, exists := properties["Tags"]
	if !exists {
		return tagsMap, nil
	}

	tagsList, ok := literalTags.([]interface{})
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
