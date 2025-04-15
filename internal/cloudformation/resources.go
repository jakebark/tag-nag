package cloudformation

import (
	"fmt"
	"strings"

	"github.com/jakebark/tag-nag/internal/shared"
	"gopkg.in/yaml.v3"
)

// getResourceViolations inspects resource blocks and returns violations
func checkResourcesforTags(resourcesMapping map[string]*yaml.Node, requiredTags shared.TagMap, caseInsensitive bool, fileLines []string, skipAll bool) []Violation {
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

		missing := shared.FilterMissingTags(requiredTags, tags, caseInsensitive)
		if len(missing) > 0 {
			violation := Violation{
				resourceName: resourceName,
				resourceType: resourceType,
				line:         resourceNode.Line,
				missingTags:  missing,
			}
			// if file-level or resource-level ignore is found
			if skipAll || skipResource(resourceNode, fileLines) {
				violation.skip = true
			}
			violations = append(violations, violation)
		}
	}
	return violations
}

// extractTagMap extracts a yaml/json map to a go map
func extractTagMap(properties map[string]interface{}, caseInsensitive bool) (shared.TagMap, error) {
	tagsMap := make(shared.TagMap)
	literalTags, exists := properties["Tags"]
	if !exists {
		return tagsMap, nil
	}

	tagsList, ok := literalTags.([]interface{})
	if !ok {
		return tagsMap, fmt.Errorf("Tags format is invalid") // tags are not in a list
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
		var tagValue string
		if valStr, ok := tagEntry["Value"].(string); ok {
			tagValue = valStr
		} else if refMap, ok := tagEntry["Value"].(map[string]interface{}); ok {
			if ref, exists := refMap["Ref"]; exists {
				if refStr, ok := ref.(string); ok {
					tagValue = fmt.Sprintf("!Ref %s", refStr)
				}
			}
		}
		if caseInsensitive {
			key = strings.ToLower(key)
		}
		tagsMap[key] = []string{tagValue}
	}
	return tagsMap, nil
}
