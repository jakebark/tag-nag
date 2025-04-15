package shared

import (
	"fmt"
	"strings"
)

func PrintViolations(filePath string, violations []Violation) {
	if len(violations) == 0 {
		return
	}
	fmt.Printf("\nViolation(s) in %s\n", filePath)
	for _, v := range violations {
		if v.Skip {
			fmt.Printf("  %d: %s \"%s\" skipped\n", v.Line, v.ResourceType, v.ResourceName)
		} else {
			fmt.Printf("  %d: %s \"%s\" 🏷️  Missing tags: %s\n", v.Line, v.ResourceType, v.ResourceName, strings.Join(v.MissingTags, ", "))
		}
	}
}

// FilterMissingTags checks effectiveTags against requiredTags
func FilterMissingTags(requiredTags TagMap, effectiveTags TagMap, caseInsensitive bool) []string {
	var missingTags []string

	// loop through key values in requiredTags
	for reqKey, allowedValues := range requiredTags {
		var effectiveValues []string
		tagFound := false

		// required key exists in effective tags
		for effKey, values := range effectiveTags {
			keyMatch := false
			if caseInsensitive {
				if strings.EqualFold(effKey, reqKey) {
					keyMatch = true
				}
			} else {
				if effKey == reqKey {
					keyMatch = true
				}
			}

			if keyMatch {
				effectiveValues = values
				tagFound = true
				break // Found the key, no need to check further keys
			}
		}

		// if no effective value is found
		if !tagFound {
			if len(allowedValues) > 0 {
				missingTags = append(missingTags, fmt.Sprintf("%s[%s]", reqKey, strings.Join(allowedValues, ",")))
			} else {
				missingTags = append(missingTags, reqKey)
			}
			continue
		}

		// if a value is found, check it is allowed (against requiredTags)
		if len(allowedValues) > 0 {
			valueMatchFound := false
			for _, allowed := range allowedValues {
				for _, effVal := range effectiveValues {
					valMatch := false
					if caseInsensitive {
						if strings.EqualFold(effVal, allowed) {
							valMatch = true
						}
					} else {
						if effVal == allowed {
							valMatch = true
						}
					}
					if valMatch {
						valueMatchFound = true
						break
					}
				}
				if valueMatchFound {
					break
				}
			}
			// effective values dont match allowed values
			if !valueMatchFound {
				missingTags = append(missingTags, fmt.Sprintf("%s[%s]", reqKey, strings.Join(allowedValues, ",")))
			}
		}
	}

	return missingTags
}
