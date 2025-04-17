package util

import (
	"strings"
)

// List of subscription topic patterns
var subList = []string{
	"devices/telemetry/control/{device_number}",   // Subscribe to control commands from the platform
	"devices/telemetry/control/{device_number}/+", // Subscribe to control commands from the platform
	"devices/attributes/set/{device_number}/+",    // Subscribe to attribute setting commands from the platform
	"devices/attributes/get/{device_number}",      // Subscribe to attribute query requests from the platform
	"devices/command/{device_number}/+",           // Subscribe to commands from the platform

	"ota/devices/infrom/{device_number}", // Receive OTA upgrade tasks (firmware-related)

	"devices/attributes/response/{device_number}/+", // Subscribe to attribute response from the platform
	"devices/event/response/{device_number}/+",      // Receive event response from the platform

	"gateway/telemetry/control/{device_number}", // Subscribe to control commands from the platform (gateway)
	"gateway/attributes/set/{device_number}/+",  // Subscribe to attribute setting commands (gateway)
	"gateway/attributes/get/{device_number}",    // Subscribe to attribute query requests (gateway)
	"gateway/command/{device_number}/+",         // Subscribe to commands (gateway)

	"gateway/attributes/response/{device_number}/+", // Subscribe to attribute response (gateway)
	"gateway/event/response/{device_number}/+",      // Receive event response (gateway)

	"{device_number}/down", // Downlink data for SmartMind integrated sprinkler device

	"devices/register/response/+",    // Response from platform to gateway sub-device registration
	"devices/config/down/response/+", // Response from platform to device configuration download
}

// ValidateSubTopic checks whether a topic matches any pattern in subList
func ValidateSubTopic(topic string) bool {
	for _, pattern := range subList {
		if matchesPatternSub(topic, pattern) {
			return true
		}
	}
	return false
}

// matchesPatternSub checks whether a topic matches a specific pattern
func matchesPatternSub(topic, pattern string) bool {
	topicParts := strings.Split(topic, "/")
	patternParts := strings.Split(pattern, "/")

	// If the number of segments is different, it's not a match
	if len(topicParts) != len(patternParts) {
		return false
	}

	// Compare each segment
	for i := range topicParts {
		switch patternParts[i] {
		case "{device_number}":
			// {device_number} should not be "+" or "#"
			if topicParts[i] == "+" || topicParts[i] == "#" {
				return false
			}
		case "+":
			// "+" can match anything except "#"
			if topicParts[i] == "#" {
				return false
			}
		default:
			// Must be an exact match
			if topicParts[i] != patternParts[i] {
				return false
			}
		}
	}

	return true
}
