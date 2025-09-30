package emit

import (
	"encoding/json"
	"sort"
)

// CanonicalJSONMarshaler provides stable JSON marshaling with sorted keys
type CanonicalJSONMarshaler struct{}

// Marshal produces canonical JSON with sorted keys for stable comparison
func (m *CanonicalJSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	// First marshal to a map to get all fields
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	// Recursively sort keys
	sortedObj := m.sortKeys(obj)

	// Marshal with sorted keys
	return json.Marshal(sortedObj)
}

// sortKeys recursively sorts map keys for stable JSON output
func (m *CanonicalJSONMarshaler) sortKeys(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		// Create a new map with sorted keys
		sorted := make(map[string]interface{})

		// Get all keys and sort them
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Recursively sort nested objects
		for _, k := range keys {
			sorted[k] = m.sortKeys(v[k])
		}

		return sorted

	case []interface{}:
		// Recursively sort array elements
		sorted := make([]interface{}, len(v))
		for i, item := range v {
			sorted[i] = m.sortKeys(item)
		}
		return sorted

	default:
		// Primitive types don't need sorting
		return v
	}
}

// Default canonical marshaler instance
var CanonicalMarshaler = &CanonicalJSONMarshaler{}
