package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTT broker address
var BROKER string = "127.0.0.1:1883"

// Message sending interval in seconds
var LOOP_TIME int = 5

// Number of simulated devices
var DEVICE_NUM int = 1000

// MQTT topic to publish
var TOPIC string = "device/attributes"

// Payload to send
var PAYLOAD string = `{"temperature": 20, "humidity": 60}`

// Quality of Service level
var QOS int = 0

func main() {
	fmt.Println("Starting script execution...")
	// Continuously create MQTT clients and let each client publish messages in a loop
	MqttPublishLoopClient(TOPIC, PAYLOAD, QOS)
}

// Create a new MQTT client connection
func MqttClient(clientId string) (mqtt.Client, error) {
	// Reconnection logic on connection loss
	var connectLostHandler mqtt.ConnectionLostHandler = func(c mqtt.Client, err error) {
		fmt.Printf("("+clientId+") MQTT connection lost: %v", err)
		i := 0
		for {
			time.Sleep(5 * time.Second)
			if !c.IsConnectionOpen() {
				i++
				fmt.Println("("+clientId+") Attempting to reconnect...", i)
				if token := c.Connect(); token.Wait() && token.Error() != nil {
					fmt.Println("(" + clientId + ") MQTT reconnection failed...")
				} else {
					break
				}
			} else {
				break
			}
		}
	}

	opts := mqtt.NewClientOptions()
	opts.SetClientID(clientId)
	opts.AddBroker(BROKER)
	opts.SetAutoReconnect(true)
	opts.SetOrderMatters(false)
	opts.OnConnectionLost = connectLostHandler
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		fmt.Println("MQTT client connected (" + clientId + ")")
	})

	reconnectAttempts := 0
	c := mqtt.NewClient(opts)

	// Asynchronously attempt to connect, retry if failed
	for {
		if token := c.Connect(); token.Wait() && token.Error() != nil {
			reconnectAttempts++
			fmt.Println("Connection error:", token.Error().Error())
			fmt.Println("MQTT client connection failed ("+clientId+")... retrying", reconnectAttempts)
		} else {
			MqttPublishLoop(TOPIC, PAYLOAD, QOS, c)
			fmt.Println("MQTT client successfully connected (" + clientId + ")")
			break
		}
		time.Sleep(5 * time.Second)
	}
	return c, nil
}

// Publish a single MQTT message
func MqttPublish(topic string, payload string, qos int, c mqtt.Client) {
	cc := c.OptionsReader()
	token := c.Publish(topic, byte(qos), false, payload)
	token.Wait()
	fmt.Printf("%s successfully sent message, topic: %s, payload: %s\n", cc.ClientID(), topic, payload)
}

// Continuously publish MQTT messages
func MqttPublishLoop(topic string, payload string, qos int, c mqtt.Client) {
	for {
		MqttPublish(topic, payload, qos, c)
		time.Sleep(time.Duration(LOOP_TIME) * time.Second)
	}
}

// Continuously create MQTT clients and send messages in loop
func MqttPublishLoopClient(topic string, payload string, qos int) {
	// Generate multiple clientIds
	for i := 0; i < DEVICE_NUM; i++ {
		clientId := fmt.Sprintf("client_%d", i)
		go MqttClient(clientId)
	}
	time.Sleep(100 * time.Second)
}
