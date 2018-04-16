package main

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/jurgen-kluft/go-home/config"
	logpkg "github.com/jurgen-kluft/go-home/logging"
	"github.com/jurgen-kluft/go-home/pubsub"
)

// Configs holds all the config objects that we can have
type Context struct {
	log     *logpkg.Logger
	pubsub  *pubsub.Context
	configs *Configs
}

type Configs struct {
	Configurations map[string]*Configuration `json:"configurations"`
}

type Configuration struct {
	Name           string `json:"name"`
	ConfigFilename string `json:"filename"`
	ChannelName    string `json:"channel"`
	ReflectType    reflect.Type
}

func ConfigFromJSON(jsonstr string) (*Configs, error) {
	c := &Configs{}
	err := json.Unmarshal([]byte(jsonstr), c)
	return c, err
}

func (c *Context) InitializeReflectTypes() {
	for name, configuration := range c.configs.Configurations {
		switch name {
		case "aqi":
			configuration.ReflectType = reflect.TypeOf(config.AqiConfig{})
		case "automation":

		case "bravia.tv":
			configuration.ReflectType = reflect.TypeOf(config.BraviaTVConfig{})
		case "calendar":
			configuration.ReflectType = reflect.TypeOf(config.CalendarConfig{})
		case "flux":
			configuration.ReflectType = reflect.TypeOf(config.FluxConfig{})
		case "hue":
			configuration.ReflectType = reflect.TypeOf(config.HueConfig{})
		case "presence":
			configuration.ReflectType = reflect.TypeOf(config.PresenceConfig{})
		case "samsung.tv":
			configuration.ReflectType = reflect.TypeOf(config.SamsungTVConfig{})
		case "shout":
			configuration.ReflectType = reflect.TypeOf(config.ShoutConfig{})
		case "suncalc":
			configuration.ReflectType = reflect.TypeOf(config.SuncalcConfig{})
		case "weather":
			configuration.ReflectType = reflect.TypeOf(config.WeatherConfig{})
		case "wemo":
			configuration.ReflectType = reflect.TypeOf(config.WemoConfig{})
		case "xiaomi":
			configuration.ReflectType = reflect.TypeOf(config.XiaomiConfig{})
		case "yee":
			configuration.ReflectType = reflect.TypeOf(config.YeeConfig{})
		}
	}
}

// SendConfigOnChannel will load the JSON based config file and publish it onto pubsub
func (c *Context) SendConfigOnChannel(configtype string, client *pubsub.Context) (err error) {
	configuration, exists := c.configs.Configurations[configtype]
	if exists {
		var data []byte
		data, err = ioutil.ReadFile(configuration.ConfigFilename)
		if err != nil {
			return err
		}
		v := reflect.New(configuration.ReflectType).Elem().Interface().(config.Config)
		v, err = v.FromJSON(string(data))
		if err == nil {
			jsonstr, err := v.ToJSON()
			if err == nil {
				err = client.Publish(configuration.ChannelName, jsonstr)
			}
		}
	}
	return
}

func main() {
	ctx := &Context{}

	ctx.log = logpkg.New("configs")
	ctx.log.AddEntry("emitter")
	ctx.log.AddEntry("configs")

	for {
		connected := true
		for connected {
			client := pubsub.New(config.EmitterSecrets["host"])
			register := []string{"config/config/"}
			subscribe := []string{"config/config/"}
			err := client.Connect("configs", register, subscribe)
			if err == nil {
				ctx.log.LogInfo("emitter", "connected")

				for connected {
					select {
					case msg := <-client.InMsgs:
						topic := msg.Topic()
						if topic == "client/disconnected/" {
							ctx.log.LogInfo("emitter", "disconnected")
							connected = false
						} else if topic == "config/config/" {
							config, err := ConfigFromJSON(string(msg.Payload()))
							if err == nil {
								ctx.log.LogInfo("configs", "received configuration")
								ctx.configs = config
								ctx.InitializeReflectTypes()
							} else {
								ctx.log.LogError("configs", err.Error())
							}
						}
						break
					case <-time.After(time.Second * 10):

						// Any config files updated ?

						break
					}
				}
			} else {
				connected = false
			}

			if err != nil {
				ctx.log.LogError("configs", err.Error())
			}
		}

		time.Sleep(5 * time.Second)
	}
}
