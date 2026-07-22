package config

func cloneTag(tag TagConfig) TagConfig {
	tag.Groups = append([]string(nil), tag.Groups...)
	tag.Values = cloneValues(tag.Values)
	tag.Suggest = cloneRelationships(tag.Suggest)
	tag.Demand = cloneRelationships(tag.Demand)
	tag.Conflict = cloneRelationships(tag.Conflict)
	return tag
}

func cloneValues(values []PredefinedValue) []PredefinedValue {
	cloned := make([]PredefinedValue, len(values))
	for index, value := range values {
		value.Suggest = cloneRelationships(value.Suggest)
		value.Demand = cloneRelationships(value.Demand)
		value.Conflict = cloneRelationships(value.Conflict)
		cloned[index] = value
	}
	return cloned
}

func cloneRelationships(relationships []Relationship) []Relationship {
	cloned := make([]Relationship, len(relationships))
	for index, relationship := range relationships {
		relationship.Has = append([]any(nil), relationship.Has...)
		relationship.Not = append([]any(nil), relationship.Not...)
		cloned[index] = relationship
	}
	return cloned
}
