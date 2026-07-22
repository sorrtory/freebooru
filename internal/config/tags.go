package config

// TagType determines how a tag value is stored and compared.
type TagType string

// Supported MVP tag types.
const (
	TagTypeBool       TagType = "bool"
	TagTypeText       TagType = "text"
	TagTypeInt        TagType = "int"
	TagTypeDate       TagType = "date"
	TagTypeDatetime   TagType = "datetime"
	TagTypeValue      TagType = "value"
	TagTypeMultivalue TagType = "multivalue"
)

// TagConfig defines one tag and its optional predefined values and rules.
type TagConfig struct {
	Name     string            `yaml:"name"`
	Type     TagType           `yaml:"type"`
	Comment  string            `yaml:"comment,omitempty"`
	Groups   []string          `yaml:"groups,omitempty"`
	Values   []PredefinedValue `yaml:"values,omitempty"`
	Suggest  []Relationship    `yaml:"suggest,omitempty"`
	Demand   []Relationship    `yaml:"demand,omitempty"`
	Conflict []Relationship    `yaml:"conflict,omitempty"`
}

// PredefinedValue declares one allowed value and rules activated by it.
type PredefinedValue struct {
	Val      string         `yaml:"val"`
	Comment  string         `yaml:"comment,omitempty"`
	Aliases  []string       `yaml:"aliases,omitempty"`
	Suggest  []Relationship `yaml:"suggest,omitempty"`
	Demand   []Relationship `yaml:"demand,omitempty"`
	Conflict []Relationship `yaml:"conflict,omitempty"`
}

// Relationship is the decoded shape of a directed tag rule.
// Predicate compatibility is checked when the relationship graph is built.
type Relationship struct {
	Tag    string `yaml:"tag"`
	Has    []any  `yaml:"has,omitempty"`
	Is     any    `yaml:"is,omitempty"`
	Not    []any  `yaml:"not,omitempty"`
	Min    *int64 `yaml:"min,omitempty"`
	Max    *int64 `yaml:"max,omitempty"`
	Before string `yaml:"before,omitempty"`
	After  string `yaml:"after,omitempty"`
	Regex  string `yaml:"regex,omitempty"`
	Reason string `yaml:"reason,omitempty"`
}
