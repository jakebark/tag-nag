package cloudformation

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// getResourceViolations inspects resource blocks and returns violations
func checkResourcesforTags(resourcesMapping map[string]*yaml.Node, requiredTags TagMap, caseInsensitive bool, fileLines []string, skipAll bool) []Violation {
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
func extractTagMap(properties map[string]interface{}, caseInsensitive bool) (TagMap, error) {
	tagsMap := make(TagMap)
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

func filterMissingTags(requiredTags TagMap, effectiveTags TagMap, caseInsensitive bool) []string {
	var missing []string
	for reqKey, allowedValues := range requiredTags {
		var effectiveValues []string
		for key, values := range effectiveTags {
			if caseInsensitive {
				if strings.EqualFold(key, reqKey) {
					effectiveValues = values
					break
				}
			} else {
				if key == reqKey {
					effectiveValues = values
					break
				}
			}
		}

		if len(effectiveValues) == 0 {
			if len(allowedValues) > 0 {
				missing = append(missing, fmt.Sprintf("%s[%s]", reqKey, strings.Join(allowedValues, ",")))
			} else {
				missing = append(missing, reqKey)
			}
			continue
		}

		if len(allowedValues) > 0 {
			var matchFound bool
			for _, allowed := range allowedValues {
				for _, effVal := range effectiveValues {
					if caseInsensitive {
						if strings.EqualFold(effVal, allowed) {
							matchFound = true
							break
						}
					} else {
						if effVal == allowed {
							matchFound = true
							break
						}
					}
				}
				if matchFound {
					break
				}
			}
			if !matchFound {
				missing = append(missing, fmt.Sprintf("%s[%s]", reqKey, strings.Join(allowedValues, ",")))
			}
		}
	}
	return missing
}
