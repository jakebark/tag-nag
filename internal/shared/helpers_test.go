package shared

import (
	"reflect"
	"sort"
	"testing"
)

func testFilterMissingTags(t *testing.T) {
	testCases := []struct {
		name            string
		requiredTags    TagMap
		effectiveTags   TagMap
		caseInsensitive bool
		expectedMissing []string
	}{
		{
			name:            "tags present",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod", "Dev"}},
			effectiveTags:   TagMap{"Owner": {"a"}, "Env": {"Prod"}},
			caseInsensitive: false,
			expectedMissing: nil,
		},
		{
			name:            "missing key",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod"}},
			effectiveTags:   TagMap{"Env": {"Prod"}},
			caseInsensitive: false,
			expectedMissing: []string{"Owner"},
		},
		{
			name:            "wrong value",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod"}},
			effectiveTags:   TagMap{"Owner": {"a"}, "Env": {"Dev"}},
			caseInsensitive: false,
			expectedMissing: []string{"Env[Prod]"},
		},
		{
			name:            "missing key and value",
			requiredTags:    TagMap{"Env": {"Prod"}},
			effectiveTags:   TagMap{"Owner": {"a"}},
			caseInsensitive: false,
			expectedMissing: []string{"Env[Prod]"},
		},
		{
			name:            "tags present, case insensitive",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod", "Dev"}},
			effectiveTags:   TagMap{"owner": {"a"}, "env": {"prod"}},
			caseInsensitive: true,
			expectedMissing: nil,
		},
		{
			name:            "missing key, case insensitive",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod"}},
			effectiveTags:   TagMap{"env": {"Prod"}},
			caseInsensitive: true,
			expectedMissing: []string{"Owner"},
		},
		{
			name:            "wrong value, case insensitive",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod"}},
			effectiveTags:   TagMap{"owner": {"a"}, "env": {"Dev"}},
			caseInsensitive: true,
			expectedMissing: []string{"Env[Prod]"},
		},
		{
			name:            "missing key and value, case insensitive",
			requiredTags:    TagMap{"Env": {"Prod"}},
			effectiveTags:   TagMap{"owner": {"a"}},
			caseInsensitive: true,
			expectedMissing: []string{"Env[Prod]"},
		},
		{
			name:            "no required tags",
			requiredTags:    TagMap{},
			effectiveTags:   TagMap{"Owner": {"a"}, "Env": {"Dev"}},
			caseInsensitive: false,
			expectedMissing: nil,
		},
		{
			name:            "no tags",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod"}},
			effectiveTags:   TagMap{},
			caseInsensitive: false,
			expectedMissing: []string{"Owner", "Env[Prod]"},
		},
		{
			name:            "multiple values required, one present",
			requiredTags:    TagMap{"Region": {"us-east-1", "us-west-2"}},
			effectiveTags:   TagMap{"Region": {"us-west-2"}},
			caseInsensitive: false,
			expectedMissing: nil,
		},
		{
			name:            "multiple values required, none present",
			requiredTags:    TagMap{"Region": {"us-east-1", "us-west-2"}},
			effectiveTags:   TagMap{"Region": {"eu-central-1"}},
			caseInsensitive: false,
			expectedMissing: []string{"Region[us-east-1,us-west-2]"},
		},
		{
			name:            "multiple values required, one present, case insensitive",
			requiredTags:    TagMap{"Env": {"Prod", "Dev"}},
			effectiveTags:   TagMap{"Env": {"prod"}},
			caseInsensitive: true,
			expectedMissing: nil,
		},
		{
			name:            "multiple values required, none present, case insensitive",
			requiredTags:    TagMap{"Region": {"us-east-1", "us-west-2"}},
			effectiveTags:   TagMap{"region": {"eu-central-1"}},
			caseInsensitive: true,
			expectedMissing: []string{"Region[us-east-1,us-west-2]"},
		},
		{
			name:            "multiple values required, multiple values present",
			requiredTags:    TagMap{"Env": {"Prod", "Dev"}},
			effectiveTags:   TagMap{"Env": {"Prod", "Stage"}},
			caseInsensitive: false,
			expectedMissing: nil,
		},
		{
			name:            "multiple values present, none match",
			requiredTags:    TagMap{"Env": {"Prod"}},
			effectiveTags:   TagMap{"Env": {"Dev", "Stage"}},
			caseInsensitive: false,
			expectedMissing: []string{"Env[Prod]"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := FilterMissingTags(tc.requiredTags, tc.effectiveTags, tc.caseInsensitive)

			if actual != nil && tc.expectedMissing != nil {
				sort.Strings(actual)
				sort.Strings(tc.expectedMissing)
			}

			if !reflect.DeepEqual(actual, tc.expectedMissing) {
				t.Errorf("FilterMissingTags() = %#v; want %#v", actual, tc.expectedMissing)
			}
		})
	}
}
