package config

import (
	"encoding/json"
	"time"
)

func WeatherConfigFromJSON(jsonstr string) (*WeatherConfig, error) {
	var r *WeatherConfig
	err := json.Unmarshal([]byte(jsonstr), r)
	return r, err
}

func (r *WeatherConfig) FromJSON() ([]byte, error) {
	return json.Marshal(r)
}

type WeatherConfig struct {
	Location    Geo         `json:"location"`
	Key         CryptString `json:"key"`
	Notify      []Notify    `json:"notify"`
	Clouds      []WItem     `json:"clouds"`
	Rain        []WItem     `json:"rain"`
	Wind        []WItem     `json:"wind"`
	Temperature []WItem     `json:"temperature"`
}

type WItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Unit        Unit   `json:"unit"`
	Range       MinMax `json:"range"`
}

type Notify struct {
	Type    string `json:"type"`
	Warning MinMax `json:"warning"`
	Alert   MinMax `json:"alert"`
}

type Unit string

const (
	Celcius Unit = "Celcius"
	Empty   Unit = "%"
	KMH     Unit = "km/h"
)

type Forecast struct {
	From        time.Time `json:"from"`
	Until       time.Time `json:"until"`
	Rain        float64   `json:"rain"`
	RainDescr   string    `json:"rainDescr"`
	Wind        float64   `json:"wind"`
	WindDescr   string    `json:"windDescr"`
	Clouds      float64   `json:"clouds"`
	CloudDescr  string    `json:"cloudsDescr"`
	Temperature float64   `json:"temperature"`
	TempDescr   string    `json:"temperatureDescr"`
}

func WeatherForecastFromJSON(jsonstr string) (*Forecast, error) {
	var r *Forecast
	err := json.Unmarshal([]byte(jsonstr), r)
	return r, err
}

func (r *Forecast) FromJSON() ([]byte, error) {
	return json.Marshal(r)
}
