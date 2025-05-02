package cloudformation

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type cfnSpec struct {
	ResourceTypes map[string]cfnResourceType `json:"ResourceTypes"`
	PropertyTypes map[string]cfnPropertyType `json:"PropertyTypes"`
}

type cfnResourceType struct {
	Properties map[string]cfnProperty `json:"Properties"`
}

type cfnProperty struct {
	Required          bool   `json:"Required"`
	Type              string `json:"Type"`     // list
	ItemType          string `json:"ItemType"` // tag
	PrimitiveType     string `json:"PrimitiveType"`
	PrimitiveItemType string `json:"PrimitiveItemType"`
	// Other fields ignored
}

type cfnPropertyType struct {
	Properties map[string]cfnProperty `json:"Properties"`
}

// isTaggable checks the cfn spec file to see if a resource can be tagged
func (rt cfnResourceType) isTaggable(specData *cfnSpec) bool {
	tagsProp, ok := rt.Properties["Tags"]
	if !ok {
		return false
	}
	if tagsProp.Type == "List" && tagsProp.ItemType == "Tag" {
		tagTypeDef, tagTypeExists := specData.PropertyTypes["Tag"]
		if tagTypeExists {
			_, keyExists := tagTypeDef.Properties["Key"]
			_, valueExists := tagTypeDef.Properties["Value"]
			return keyExists && valueExists
		}
	}
	return false
}

// LoadTaggableResourcesFromSpec parses the provided spec file path
func loadTaggableResourcesFromSpec(specFilePath string) (map[string]bool, error) {
	log.Printf("Attempting to load CloudFormation specification from: %s", specFilePath)
	data, err := os.ReadFile(specFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CloudFormation spec file '%s': %w", specFilePath, err)
	}

	var specData cfnSpec
	err = json.Unmarshal(data, &specData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CloudFormation spec JSON from '%s': %w", specFilePath, err)
	}

	if specData.ResourceTypes == nil {
		return nil, fmt.Errorf("invalid CloudFormation spec format: missing 'ResourceTypes' in '%s'", specFilePath)
	}
	if specData.PropertyTypes == nil {
		return nil, fmt.Errorf("invalid CloudFormation spec format: missing 'PropertyTypes' in '%s'", specFilePath)
	}

	taggableMap := make(map[string]bool)
	for resourceName, resourceDef := range specData.ResourceTypes {
		if strings.HasPrefix(resourceName, "AWS::") {
			taggableMap[resourceName] = resourceDef.isTaggable(&specData)
		}
	}

	log.Printf("Successfully loaded taggability info for %d AWS resource types from spec.", len(taggableMap))
	return taggableMap, nil
}
