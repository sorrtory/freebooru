package config

import "sort"

// buildGroups derives navigation groups from valid indexed tags. Sorting the
// tag keys makes group display spelling and membership order deterministic.
func buildGroups(tags map[string]catalogEntry[TagConfig]) map[string]Group {
	tagKeys := make([]string, 0, len(tags))
	for key := range tags {
		tagKeys = append(tagKeys, key)
	}
	sort.Strings(tagKeys)

	groups := make(map[string]Group)
	for _, tagKey := range tagKeys {
		for _, name := range tags[tagKey].value.Groups {
			key := normalizeName(name)
			group, ok := groups[key]
			if !ok {
				group.Name = name
			}
			group.tags = append(group.tags, tagKey)
			groups[key] = group
		}
	}
	return groups
}
