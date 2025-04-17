package thingsly

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/DrmagicE/gmqtt/config"
	"github.com/DrmagicE/gmqtt/server"
	"github.com/spf13/viper"
)

var _ server.Plugin = (*Thingsly)(nil)

const Name = "thingsly"

func init() {
	log.Println("Initializing system configuration file...")
	viper.SetEnvPrefix("GMQTT")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName("thingsly")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("failed to read configuration file: %s", err))
	}
	log.Println("System configuration file initialization completed")
	Init() // Start database and Redis
	go DefaultMqttClient.MqttInit()
	server.RegisterPlugin(Name, New)
	config.RegisterDefaultPluginConfig(Name, &DefaultConfig)
}

func New(config config.Config) (server.Plugin, error) {
	//panic("implement me")
	return &Thingsly{}, nil
}

var Log *zap.Logger

type Thingsly struct {
}

func (t *Thingsly) Load(service server.Server) error {
	Log = server.LoggerWithField(zap.String("plugin", Name))
	return nil
}

func (t *Thingsly) Unload() error {
	return nil
}

func (t *Thingsly) Name() string {
	return Name
}

func (t *Thingsly) UpdateStatus(accessToken string, status string) {
	url := "/api/device/status"
	method := "POST"
	payload := strings.NewReader(`"accessToken": "` + accessToken + `","values":{"status": "` + status + `"}}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
