package cloudformation

import (
	"encoding/json"
	"testing"
)

// helper to create a test spec file
func createTestSpec() *cfnSpec {
	specJSON := `{
		"PropertyTypes": {
			"Tag": {
				"Properties": {
					"Key": { "PrimitiveType": "String", "Required": true },
					"Value": { "PrimitiveType": "String", "Required": true }
				}
			},
			"OtherType": {
				"Properties": { "Name": { "PrimitiveType": "String" } }
			}
		},
		"ResourceTypes": {}
	}`
	var spec cfnSpec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		panic("Failed to unmarshal base test spec data: " + err.Error())
	}
	return &spec
}

func testIsTaggable(t *testing.T) {
	baseSpec := createTestSpec()

	tests := []struct {
		name         string
		resourceJSON string
		want         bool
	}{
		{
			name: "taggable",
			resourceJSON: `{
				"Properties": {
					"Name": { "PrimitiveType": "String" },
					"Tags": { "Type": "List", "ItemType": "Tag", "Required": false }
				}
			}`,
			want: true,
		},
		{
			name: "not taggable",
			resourceJSON: `{
				"Properties": {
					"Name": { "PrimitiveType": "String" }
				}
			}`,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var resourceDef cfnResourceType
			if err := json.Unmarshal([]byte(tc.resourceJSON), &resourceDef); err != nil {
				t.Fatalf("Failed to unmarshal test resource JSON: %v", err)
			}

			if got := resourceDef.isTaggable(baseSpec); got != tc.want {
				t.Errorf("isTaggable() = %v, want %v", got, tc.want)
			}
		})
	}
}
