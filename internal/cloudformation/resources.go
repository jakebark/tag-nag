package cloudformation

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// checkResourcesforTags inspects resource blocks and returns violations
func checkResourcesforTags(resourcesMapping map[string]*yaml.Node, requiredTags TagMap, caseInsensitive bool) []Violation {
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
				resourceName: resourceName,
				resourceType: resourceType,
				line:         resourceNode.Line,
				missingTags:  missing,
			})
		}
	}
	return violations
}

// extractTagMap extracts a yaml/json map to a go map
func extractTagMap(properties map[string]interface{}, caseInsensitive bool) (TagMap, error) {
	tagsMap := make(map[string]string)
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
		} else {
			if refMap, ok := tagEntry["Value"].(map[string]interface{}); ok {
				if ref, exists := refMap["Ref"]; exists {
					if refStr, ok := ref.(string); ok {
						tagValue = fmt.Sprintf("!Ref %s", refStr)
					}
				}
			}
		}
		if caseInsensitive {
			key = strings.ToLower(key)
		}
		tagsMap[key] = tagValue
	}
	return tagsMap, nil
}

func filterMissingTags(requiredTags TagMap, effectiveTags TagMap, caseInsensitive bool) []string {
	var missing []string
	for reqKey, reqVal := range requiredTags {
		found := false
		for key, value := range effectiveTags {
			if caseInsensitive {
				if !strings.EqualFold(key, reqKey) {
					continue
				}
				if reqVal != "" && !strings.EqualFold(value, reqVal) {
					continue
				}
				found = true
				break
			} else {
				if key != reqKey {
					continue
				}
				if reqVal != "" && value != reqVal {
					continue
				}
				found = true
				break
			}
		}
		if !found {
			// If a value is specified, include it in the missing output.
			if reqVal != "" {
				missing = append(missing, fmt.Sprintf("%s:%s", reqKey, reqVal))
			} else {
				missing = append(missing, reqKey)
			}
		}
	}
	return missing
}

func skipResources(violations []Violation, fileText string) (filtered, skipped []Violation) {
	lines := strings.Split(fileText, "\n")
	for _, v := range violations {
		if v.line < len(lines) {
			ignoreLine := strings.TrimSpace(lines[v.line])
			if ignoreRegex.MatchString(ignoreLine) {
				skipped = append(skipped, v)
				continue
			}
		}
		filtered = append(filtered, v)
	}
	return
}
