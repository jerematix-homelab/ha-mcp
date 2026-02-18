// Package homeassistant provides helper detection for Config Entry Flow platforms.
package homeassistant

// configEntryPlatforms defines helper platforms that require HTTP Config Entry Flow.
// These platforms cannot be created/deleted via WebSocket API and must use the
// HTTP-based Config Entry Flow mechanism instead.
var configEntryPlatforms = map[string]bool{
	"threshold":          true, // Creates binary_sensor entities
	"derivative":         true, // Creates sensor entities
	"integration":        true, // Creates sensor entities (Home Assistant's name for integral)
	"group":              true, // Creates entities matching member domain (light.*, sensor.*, etc.)
	"template":           true, // Creates sensor or binary_sensor entities
	"utility_meter":      true, // Creates sensor + select entities
	"min_max":            true, // Creates sensor entities
	"statistics":         true, // Creates sensor entities
	"trend":              true, // Creates binary_sensor entities
	"random":             true, // Creates sensor or binary_sensor entities
	"filter":             true, // Creates sensor entities
	"tod":                true, // Creates binary_sensor entities (Time of Day)
	"generic_thermostat": true, // Creates climate entities
	"switch_as_x":        true, // Creates entities based on target_domain (cover/fan/light/lock/siren/valve)
	"generic_hygrostat":  true, // Creates humidifier entities
}

// RequiresConfigEntryFlow checks if the given platform requires HTTP Config Entry Flow.
// Returns true for platforms that cannot be created/deleted via WebSocket API.
func RequiresConfigEntryFlow(platform string) bool {
	return configEntryPlatforms[platform]
}
