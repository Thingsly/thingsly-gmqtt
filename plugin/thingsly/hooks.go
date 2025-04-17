package thingsly

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DrmagicE/gmqtt/plugin/thingsly/util"
	"github.com/DrmagicE/gmqtt/server"
	"github.com/spf13/viper"
)

func (t *Thingsly) HookWrapper() server.HookWrapper {
	return server.HookWrapper{
		OnBasicAuthWrapper:  t.OnBasicAuthWrapper,
		OnSubscribeWrapper:  t.OnSubscribeWrapper,
		OnMsgArrivedWrapper: t.OnMsgArrivedWrapper,
		OnConnectedWrapper:  t.OnConnectedWrapper,
		OnClosedWrapper:     t.OnClosedWrapper,
	}
}

func (t *Thingsly) OnBasicAuthWrapper(pre server.OnBasicAuth) server.OnBasicAuth {
	return func(ctx context.Context, client server.Client, req *server.ConnectRequest) (err error) {
		err = pre(ctx, client, req)
		if err != nil {
			Log.Error(err.Error())
			return err
		}
		if string(req.Connect.Username) == "root" {
			password := viper.GetString("mqtt.password")
			if string(req.Connect.Password) == password {
				return nil
			} else {
				err := errors.New("password error;")
				Log.Warn(err.Error())
				return err
			}
		}
		if string(req.Connect.Username) == "plugin" {
			password := viper.GetString("mqtt.plugin_password")
			if string(req.Connect.Password) == password {
				return nil
			} else {
				err := errors.New("password error;")
				Log.Warn(err.Error())
				return err
			}
		}
		// ... Handle authentication logic for this plugin
		Log.Info("Auth Username: " + string(req.Connect.Username))
		Log.Info("Auth Password: " + string(req.Connect.Password))

		// voucher is a string; if there is no password, voucher is {"username":"xxx"},
		// if there is a password, voucher is {"username":"xxx","password":"xxx"}
		voucher := ""
		if string(req.Connect.Password) != "" {
			voucher = fmt.Sprintf(`{"username":"%s","password":"%s"}`, string(req.Connect.Username), string(req.Connect.Password))
		} else {
			voucher = fmt.Sprintf(`{"username":"%s"}`, string(req.Connect.Username))
		}
		// Verify the device using the voucher
		Log.Debug("voucher: " + voucher)
		device, err := GetDeviceByVoucher(voucher)
		if err != nil {
			Log.Warn(err.Error())
			return err
		}
		Log.Info("Device Voucher: " + device.Voucher)
		Log.Info("ClientID: " + string(req.Connect.ClientID))
		// MQTT client ID must be unique
		err = SetStr("mqtt_client_id_"+string(req.Connect.ClientID), device.ID, 0)
		if err != nil {
			Log.Warn(err.Error())
			return err
		}
		return nil
	}
}

func (t *Thingsly) OnConnectedWrapper(pre server.OnConnected) server.OnConnected {
	return func(ctx context.Context, client server.Client) {
		// After client connects
		// Topic: device/status
		// Payload: {"token":username,"SYS_STATUS":"online"}
		// username is the client's username

		if client.ClientOptions().Username != "root" && client.ClientOptions().Username != "plugin" {
			deviceId, err := GetStr("mqtt_client_id_" + client.ClientOptions().ClientID)
			if err != nil {
				Log.Warn("Failed to get device ID")
				return
			}
			if deviceId == "" {
				Log.Warn("Device ID does not exist")
				return
			}
			if err := DefaultMqttClient.SendData("devices/status/"+deviceId, []byte("1")); err != nil {
				Log.Warn("Failed to report status")
			}
			Log.Info("Device status sent successfully")
		}
	}
}

func (t *Thingsly) OnClosedWrapper(pre server.OnClosed) server.OnClosed {
	return func(ctx context.Context, client server.Client, err error) {
		// After client disconnects
		// Topic: device/status
		// Payload: {"token":username,"SYS_STATUS":"offline"}
		// username is the client's username

		if client.ClientOptions().Username != "root" || client.ClientOptions().Username != "plugin" {
			deviceId, err := GetStr("mqtt_client_id_" + client.ClientOptions().ClientID)
			if err != nil {
				Log.Warn("Failed to get device ID")
				return
			}
			if deviceId == "" {
				Log.Warn("Device ID does not exist")
				return
			}
			if err := DefaultMqttClient.SendData("devices/status/"+deviceId, []byte("0")); err != nil {
				Log.Warn("Failed to report status")
			}
			Log.Info("Device status sent successfully")
		}
	}
}

// Subscribe message hook function
func (t *Thingsly) OnSubscribeWrapper(pre server.OnSubscribe) server.OnSubscribe {
	return func(ctx context.Context, client server.Client, req *server.SubscribeRequest) error {
		username := client.ClientOptions().Username
		// Allow root
		if username == "root" || username == "plugin" {
			return nil
		}

		the_sub := req.Subscribe.Topics[0].Name
		// Verify the device's subscription permissions
		if !util.ValidateSubTopic(the_sub) {
			Log.Warn("Subscription permission verification failed: " + the_sub)
			return errors.New("permission denied")
		}
		Log.Info("Subscription permission verification succeeded: " + the_sub)
		return nil
	}
}

func (t *Thingsly) OnMsgArrivedWrapper(pre server.OnMsgArrived) server.OnMsgArrived {
	return func(ctx context.Context, client server.Client, req *server.MsgArrivedRequest) (err error) {
		username := client.ClientOptions().Username
		Log.Info(fmt.Sprintf("OnMsgArrivedWrapper: username %s payload %s", username, string(req.Message.Payload)))
		// Forward messages for root and plugin users directly
		if username == "root" || username == "plugin" {
			RootMessageForwardWrapper(req.Message.Topic, req.Message.Payload, false)
			return nil
		}

		the_pub := string(req.Publish.TopicName)
		// Verify the device's publish permissions
		if !util.ValidateTopic(the_pub) {
			return errors.New("permission denied")
		}

		// Allow topics ending with "/up" directly [Mindjoy-MW]
		if the_pub[len(the_pub)-3:] == "/up" {
			return nil
		}

		// Rewrite message
		newMsgMap := make(map[string]interface{})
		deviceId, err := GetStr("mqtt_client_id_" + client.ClientOptions().ClientID)
		if err != nil {
			return err
		}
		newMsgMap["device_id"] = deviceId
		newMsgMap["values"] = req.Message.Payload
		newMsgJson, _ := json.Marshal(newMsgMap)
		req.Message.Payload = newMsgJson
		// If the original topic is converted, discard the message and republish to the converted topic
		if the_pub != string(req.Publish.TopicName) {
			DefaultMqttClient.SendData(the_pub, req.Message.Payload)
			return errors.New("message is discarded;")
		}
		return nil
	}
}
