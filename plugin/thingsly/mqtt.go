package thingsly

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
)

type MqttClient struct {
	Client mqtt.Client
	IsFlag bool
}

var DefaultMqttClient *MqttClient = &MqttClient{}

func (c *MqttClient) MqttInit() error {
	opts := mqtt.NewClientOptions()
	opts.SetUsername("root")
	password := viper.GetString("mqtt.password")
	opts.SetPassword(password)
	addr := viper.GetString("mqtt.broker")
	if addr == "" {
		addr = "localhost:1883"
	}
	opts.AddBroker(addr)
	// Clean session
	opts.SetCleanSession(true)
	// Auto reconnect on failure
	opts.SetAutoReconnect(true)
	opts.SetConnectRetryInterval(1 * time.Second)   // Initial connection retry interval
	opts.SetMaxReconnectInterval(200 * time.Second) // Maximum retry interval after losing connection

	opts.SetOrderMatters(false) // Set message order
	//opts.OnConnectionLost = connectLostHandler
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		fmt.Println("Mqtt client connected")
	})
	opts.SetClientID("thingsly-gmqtt-client")
	c.Client = mqtt.NewClient(opts)
	// Wait for successful connection
	for {
		if token := c.Client.Connect(); token.Wait() && token.Error() != nil {
			fmt.Println("Mqtt client connection failed (", addr, "), waiting to reconnect...")
			time.Sleep(1 * time.Second)
		} else {
			fmt.Println("Mqtt client connected successfully")
			c.IsFlag = true
			break
		}
	}
	return nil
}

func (c *MqttClient) SendData(topic string, data []byte) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("【SendData】Exception captured:", err)
			return
		}
	}()
	//go func() {
	Log.Info("Checking MqttClient connection status...")
	if !c.IsFlag {
		i := 1
		for {
			fmt.Println("Waiting...", i)
			if i == 10 || c.IsFlag {
				break
			}
			time.Sleep(1 * time.Second)
			i++
		}
	}
	Log.Info("Sending device status...")
	token := c.Client.Publish(topic, 0, false, string(data))
	if !token.WaitTimeout(5 * time.Second) {
		Log.Warn("Sending device status timed out")
	} else if err := token.Error(); err != nil {
		Log.Warn("Failed to send device status: " + err.Error())
	}
	Log.Info("Sending device status completed")
	//}()
	return nil
}
