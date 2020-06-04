package config

import "encoding/json"

// AhkConfigFromJSON parser the incoming JSON string and returns an Config instance for Ahk
func AhkConfigFromJSON(data []byte) (*AhkConfig, error) {
	r := &AhkConfig{}
	err := json.Unmarshal(data, r)
	return r, err
}

// FromJSON converts json data into AhkConfig
func (r *AhkConfig) FromJSON(data []byte) error {
	c := AhkConfig{}
	err := json.Unmarshal(data, &c)
	*r = c
	return err
}

// ToJSON converts a AhkConfig to a JSON string
func (r *AhkConfig) ToJSON() ([]byte, error) {
	data, err := json.Marshal(r)
	if err == nil {
		return data, nil
	}
	return nil, err
}

// AhkConfig contains information for the Apple Homekit service
type AhkConfig struct {
	Pin       string        `json:"pin"`
	Subscribe []string      `json:"subscribe"`
	Register  []AhkRegister `json:"register"`
	Lights    []AhkLight    `json:"lights"`
	Switches  []AhkLight    `json:"switches"`
	Sensors   []AhkLight    `json:"sensors"`
}

type AhkLight struct {
	Name    string       `json:"name"`
	Type    *string      `json:"type,omitempty"`
	Channel *AhkRegister `json:"channel,omitempty"`
	ID      int64        `json:"id"`
}

type AhkRegister string

const (
	StateLightAhk  AhkRegister = "state/light/ahk/"
	StateSwitchAhk AhkRegister = "state/switch/ahk/"
)
