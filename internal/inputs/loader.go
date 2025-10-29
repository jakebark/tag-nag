package inputs

import (
	"fmt"
	"os"

	"github.com/jakebark/tag-nag/internal/config"
	"github.com/jakebark/tag-nag/internal/shared"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Tags     []TagDefinition `yaml:"tags"`
	Settings Settings        `yaml:"settings"`
	Skip     []string        `yaml:"skip"`
}

type TagDefinition struct {
	Key    string   `yaml:"key"`
	Values []string `yaml:"values,omitempty"`
}

type Settings struct {
	CaseInsensitive bool   `yaml:"case_insensitive"`
	DryRun          bool   `yaml:"dry_run"`
	CfnSpec         string `yaml:"cfn_spec"`
}

// FindAndLoadConfigFile attempts to find and load configuration file
func FindAndLoadConfigFile() (*Config, error) {
	if _, err := os.Stat(config.DefaultConfigFile); err == nil {
		return processConfigFile(config.DefaultConfigFile)
	}

	if _, err := os.Stat(config.AltConfigFile); err == nil {
		return processConfigFile(config.AltConfigFile)
	}

	return nil, nil
}

// processConfigFile reads the config file
func processConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	return &config, nil
}

// ConvertToTagMap converts config tags to internal TagMap format
func (c *Config) convertToTagMap() shared.TagMap {
	tagMap := make(shared.TagMap)

	for _, tag := range c.Tags {
		tagMap[tag.Key] = tag.Values
	}

	return tagMap
}
