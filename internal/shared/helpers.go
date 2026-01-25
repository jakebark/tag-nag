package shared

import (
	"fmt"
	"sort"
	"strings"
)

// FilterMissingTags checks effectiveTags against requiredTags
func FilterMissingTags(requiredTags TagMap, effectiveTags TagMap, caseInsensitive bool) []string {
	var missingTags []string

	for requiredKey, allowedValues := range requiredTags {
		effectiveValues, keyFound := matchTagKey(requiredKey, effectiveTags, caseInsensitive)

		// construct violation message
		violationMessage := requiredKey
		if len(allowedValues) > 0 {
			violationMessage = fmt.Sprintf("%s[%s]", requiredKey, strings.Join(allowedValues, ","))
		}
		if !keyFound {
			missingTags = append(missingTags, violationMessage)
			continue
		}

		// if there are tag values required, check them
		if len(allowedValues) > 0 {
			if !matchTagValue(allowedValues, effectiveValues, caseInsensitive) {
				missingTags = append(missingTags, violationMessage)
			}
		}
	}

	sort.Strings(missingTags) //
	return missingTags
}

// matchTagKey checks required tag key against effective tags
func matchTagKey(requiredKey string, effectiveTags TagMap, caseInsensitive bool) (values []string, found bool) {
	for effectiveKey, effectiveValues := range effectiveTags {
		if CompareCase(effectiveKey, requiredKey, caseInsensitive) {
			return effectiveValues, true
		}
	}
	return nil, false
}

// matchTagValue checks required tag values (if present) against effective tags
func matchTagValue(allowedValues []string, effectiveValues []string, caseInsensitive bool) bool {
	if len(allowedValues) == 0 { // if no tag alues are required, return match
		return true
	}
	if len(effectiveValues) == 0 && len(allowedValues) > 0 {
		return false
	}

	for _, allowed := range allowedValues {
		for _, effectiveValue := range effectiveValues {
			if CompareCase(effectiveValue, allowed, caseInsensitive) {
				return true
			}
		}
	}
	return false
}

// NormalizeCase lowers the case if caseInsensitive is true
func NormalizeCase(input string, caseInsensitive bool) string {
	if caseInsensitive {
		return strings.ToLower(input)
	}
	return input
}

// CompareCase compares case sensitivity where appropriate
func CompareCase(first, second string, caseInsensitive bool) bool {
	if caseInsensitive {
		return strings.EqualFold(first, second)
	}
	return first == second
}
