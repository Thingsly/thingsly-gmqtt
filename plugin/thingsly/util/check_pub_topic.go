package util

import (
	"strings"
)

// List of topic patterns for IoT device publishing
var pubList = []string{
	"devices/telemetry",                 // Telemetry reporting
	"devices/attributes/+",              // Attribute reporting
	"devices/event/+",                   // Event reporting
	"ota/devices/progress",              // Device OTA progress updates
	"devices/attributes/set/response/+", // Attribute set response reporting
	"devices/command/response/+",        // Command response reporting

	"gateway/telemetry",                 // Telemetry reporting (via gateway)
	"gateway/attributes/+",              // Attribute reporting (gateway)
	"gateway/event/+",                   // Event reporting (gateway)
	"gateway/attributes/set/response/+", // Attribute set response reporting (gateway)
	"gateway/command/response/+",        // Command response reporting (gateway)

	"devices/register",    // Registration of gateway sub-devices
	"devices/config/down", // Device configuration download

	"+/up", // Uplink data from SmartMind integrated sprinkler device
}

// MQTT single-level wildcard
const mqttWildcard = "+"

// ValidateTopic checks whether a topic matches any pattern in pubList
func ValidateTopic(topic string) bool {
	for _, pattern := range pubList {
		if matchesPattern(topic, pattern) {
			return true
		}
	}
	return false
}

// matchesPattern checks if a topic matches the specified pattern
func matchesPattern(topic, pattern string) bool {
	topicParts := strings.Split(topic, "/")
	patternParts := strings.Split(pattern, "/")

	// Return false if the number of segments is different
	if len(topicParts) != len(patternParts) {
		return false
	}

	// Check each segment for direct match or wildcard match
	for i := range topicParts {
		if patternParts[i] != mqttWildcard && topicParts[i] != patternParts[i] {
			return false
		}
	}

	return true
}
