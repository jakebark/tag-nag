package shared

import (
	"fmt"
	"sort"
	"strings"
)

// FilterMissingTags checks effectiveTags against requiredTags
func FilterMissingTags(requiredTags TagMap, effectiveTags TagMap, caseInsensitive bool) []string {
	var missingTags []string

	for reqKey, allowedValues := range requiredTags {
		effectiveValues, keyFound := matchTagKey(reqKey, effectiveTags, caseInsensitive)

		// construct violation message
		violationMessage := reqKey
		if len(allowedValues) > 0 {
			violationMessage = fmt.Sprintf("%s[%s]", reqKey, strings.Join(allowedValues, ","))
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
func matchTagKey(reqKey string, effectiveTags TagMap, caseInsensitive bool) (values []string, found bool) {
	for effKey, effValues := range effectiveTags {
		if caseInsensitive {
			if strings.EqualFold(effKey, reqKey) {
				return effValues, true
			}
		} else {
			if effKey == reqKey {
				return effValues, true
			}
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
		for _, effVal := range effectiveValues {
			match := false
			if caseInsensitive {
				if strings.EqualFold(effVal, allowed) {
					match = true
				}
			} else {
				if effVal == allowed {
					match = true
				}
			}
			if match {
				return true // found a match
			}
		}
	}
	return false // no match
}
