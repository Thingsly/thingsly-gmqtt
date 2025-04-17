package thingsly

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

// List of allowed topics for subscription
var AllowSubscribeTopicList = [2]string{
	"+/devices/${username}/sys/commands/+",       // Command issuance (Huawei Cloud IoT Platform specification)
	"+/devices/${username}/sys/properties/set/+", // Property setting (Huawei Cloud IoT Platform specification)
}

// Topic conversion rules
var TopicConvertMap = map[string]string{
	"+/devices/${username}/sys/properties/report": "device/attributes", // Property reporting (Huawei Cloud IoT Platform specification)
	"+/devices/${username}/sys/events/up":         "device/event",      // Event reporting (Huawei Cloud IoT Platform specification)
}

// Callback function when subscribing to a topic, used to check if the subscription is allowed
func OtherOnSubscribeWrapper(topic string, username string) error {
	fmt.Println("OtherOnSubscribeWrapper--" + topic)
	// Allowed topics for property reporting: +/devices/+/sys/properties/report
	// Allowed topics for event reporting: +/devices/+/sys/events
	// If the topic is in the topicList, allow subscription, note: ${} contains variables, + is the MQTT wildcard
	for _, v := range AllowSubscribeTopicList {
		// Replace ${username} with username
		v = strings.Replace(v, "${username}", username, -1)
		// Replace + with [^/]+, replace # with .*, then use regular expression to match
		reg := strings.Replace(v, "+", "[^/]+", -1)
		reg = strings.Replace(reg, "#", ".*", -1)
		match, _ := regexp.MatchString(reg, topic)
		if match {
			return nil
		} else {
			fmt.Println("topic not allowed--")
			return errors.New("topic not allowed")
		}
	}
	return errors.New("topic not allowed")
}

// Callback function when a message is received, used to check if receiving the message is allowed
// Topic conversion can be done here
func OtherOnMsgArrivedWrapper(topic string, username string) (string, error) {
	// If the topic is in the topicConvertMap, convert it to the corresponding topic
	for k, v := range TopicConvertMap {
		// Replace ${username} with username
		k = strings.Replace(k, "${username}", username, -1)
		// Replace + with [^/]+, replace # with .*, then use regular expression to match
		reg := strings.Replace(k, "+", "[^/]+", -1)
		reg = strings.Replace(reg, "#", ".*", -1)
		match, _ := regexp.MatchString(reg, topic)
		if match {
			return v, nil
		} else {
			return topic, errors.New("topic not allowed")
		}
	}
	return topic, errors.New("topic not allowed")
}

// If the root user's message topic is device/attributes/+, forward the message
// If flag is true, forward the message
func RootMessageForwardWrapper(topic string, payload []byte, flag bool) error {
	// If flag is false, return directly
	if !flag {
		return nil
	}
	var username string
	var topicConvertMap = map[string]string{
		"device/attributes/+": "mindjob/devices/${username}/sys/properties/set/request_id=", // Command issuance (Huawei Cloud IoT Platform specification)
	}
	// If the topic is one of the keys in topicConvertMap, convert it to the corresponding value
	for k, v := range topicConvertMap {
		// Replace + with [^/]+, replace # with .*, then use regular expression to match
		reg := strings.Replace(k, "+", "[^/]+", -1)
		reg = strings.Replace(reg, "#", ".*", -1)
		match, _ := regexp.MatchString(reg, topic)
		if match {
			// Get username
			username = strings.Split(topic, "/")[2]
			// Replace ${username} with username
			v = strings.Replace(v, "${username}", username, -1)
			// Generate a random 6-digit string number (instead of using RandomString)
			request_id := strconv.Itoa(rand.Intn(899999) + 100000)
			// Add request_id to v, where request_id is a random 6-digit number
			v = v + request_id
			// Forward the message
			fmt.Println("RootMessageForwardWrapper--" + v)
			if err := DefaultMqttClient.SendData(v, payload); err != nil {
				return err
			}
		}
	}
	return nil
}
