package entity

// Meta is used to provide custom configuration data to a sensor.
type Meta map[string]interface{}

func (m Meta) GetBool(key string) bool {
	if v, ok := m[key]; ok {
		return v == true
	}
	return false
}

func (m Meta) GetString(key string) string {
	if v, ok := m[key]; ok {
		if value, isString := v.(string); isString {
			return value
		}
	}
	return ""
}
