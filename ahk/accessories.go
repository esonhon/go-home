package main

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
	"github.com/jurgen-kluft/go-home/config"
)

type television struct {
	*accessory.Accessory
	Tv *service.Television
}

func (t *television) PowerSelect(state int) {

}

type coloredLightbulb struct {
	*accessory.Accessory
	Light *service.ColoredLightbulb
}

func (c *coloredLightbulb) Callback(onoff bool) {

}

type lightbulb struct {
	*accessory.Accessory
	Light *service.Lightbulb
}

func (c *lightbulb) Callback(onoff bool) {

}

type button struct {
	*accessory.Accessory
	Button *service.Switch
}

func (c *button) Callback(onoff bool) {

}

type motionSensor struct {
	*accessory.Accessory
	Sensor *service.MotionSensor
}

type lightSensor struct {
	*accessory.Accessory
	Sensor *service.LightSensor
}

type occupancySensor struct {
	*accessory.Accessory
	Sensor *service.OccupancySensor
}

type contactSensor struct {
	*accessory.Accessory
	Sensor *service.ContactSensor
}

type airQualitySensor struct {
	*accessory.Accessory
	Sensor *service.AirQualitySensor
}

func newTelevision(info accessory.Info) *television {
	acc := &television{}
	acc.Accessory = accessory.New(info, accessory.TypeTelevision)
	acc.Tv = service.NewTelevision()
	acc.AddService(acc.Tv.Service)
	return acc
}

func newColoredLightbulb(info accessory.Info) *coloredLightbulb {
	acc := &coloredLightbulb{}
	acc.Accessory = accessory.New(info, accessory.TypeLightbulb)
	acc.Light = service.NewColoredLightbulb()
	acc.Light.Brightness.SetValue(100)
	acc.AddService(acc.Light.Service)
	return acc
}
func newLightbulb(info accessory.Info) *lightbulb {
	acc := &lightbulb{}
	acc.Accessory = accessory.New(info, accessory.TypeLightbulb)
	acc.Light = service.NewLightbulb()
	acc.AddService(acc.Light.Service)
	return acc
}
func newButton(info accessory.Info) *button {
	acc := &button{}
	acc.Accessory = accessory.New(info, accessory.TypeSwitch)
	acc.Button = service.NewSwitch()
	acc.AddService(acc.Button.Service)
	return acc
}

func newMotionSensor(info accessory.Info) *motionSensor {
	acc := &motionSensor{}
	acc.Accessory = accessory.New(info, accessory.TypeSensor)
	acc.Sensor = service.NewMotionSensor()
	acc.AddService(acc.Sensor.Service)
	return acc
}

func newLightSensor(info accessory.Info) *lightSensor {
	acc := &lightSensor{}
	acc.Accessory = accessory.New(info, accessory.TypeSensor)
	acc.Sensor = service.NewLightSensor()
	acc.AddService(acc.Sensor.Service)
	return acc
}

func newOccupancySensor(info accessory.Info) *occupancySensor {
	acc := &occupancySensor{}
	acc.Accessory = accessory.New(info, accessory.TypeSensor)
	acc.Sensor = service.NewOccupancySensor()
	acc.AddService(acc.Sensor.Service)
	return acc
}

func newContactSensor(info accessory.Info) *contactSensor {
	acc := &contactSensor{}
	acc.Accessory = accessory.New(info, accessory.TypeSensor)
	acc.Sensor = service.NewContactSensor()
	acc.AddService(acc.Sensor.Service)
	return acc
}

func newAirQualitySensor(info accessory.Info) *airQualitySensor {
	acc := &airQualitySensor{}
	acc.Accessory = accessory.New(info, accessory.TypeSensor)
	acc.Sensor = service.NewAirQualitySensor()
	acc.AddService(acc.Sensor.Service)
	return acc
}

type accessories struct {
	Bridge            *accessory.Bridge
	ColoredLights     []*coloredLightbulb
	WhiteLights       []*lightbulb
	MotionSensors     []*motionSensor
	LightSensors      []*lightSensor
	OccupancySensors  []*occupancySensor
	ContactSensors    []*contactSensor
	AirQualitySensors []*airQualitySensor
	Switches          []*button
	Televisions       []*television
}

func (a *accessories) initializeFromConfig(config config.AhkConfig) []*accessory.Accessory {

	a.Bridge = accessory.NewBridge(accessory.Info{Name: "Bridge", ID: 1})
	a.ColoredLights = make([]*coloredLightbulb, 0, 10)
	a.WhiteLights = make([]*lightbulb, 0, 10)
	a.MotionSensors = make([]*motionSensor, 0, 10)
	a.LightSensors = make([]*lightSensor, 0, 10)
	a.OccupancySensors = make([]*occupancySensor, 0, 10)
	a.ContactSensors = make([]*contactSensor, 0, 10)
	a.AirQualitySensors = make([]*airQualitySensor, 0, 10)
	a.Switches = make([]*button, 0, 10)

	for _, lght := range config.Lights {
		if lght.Type == "colored" {
			lightbulb := newColoredLightbulb(accessory.Info{Name: lght.Name, ID: lght.ID})
			lightbulb.Light.On.OnValueRemoteUpdate(lightbulb.Callback)
			a.ColoredLights = append(a.ColoredLights, lightbulb)
		} else if lght.Type == "white" {
			lightbulb := newLightbulb(accessory.Info{Name: lght.Name, ID: lght.ID})
			lightbulb.Light.On.OnValueRemoteUpdate(lightbulb.Callback)
			a.WhiteLights = append(a.WhiteLights, lightbulb)
		}
	}

	for _, ms := range config.Sensors {
		if ms.Type == "motion" {
			sensor := newMotionSensor(accessory.Info{Name: ms.Name, ID: ms.ID})
			a.MotionSensors = append(a.MotionSensors, sensor)
		} else if ms.Type == "contact" {
			sensor := newContactSensor(accessory.Info{Name: ms.Name, ID: ms.ID})
			a.ContactSensors = append(a.ContactSensors, sensor)
		} else if ms.Type == "air-quality" {
			sensor := newAirQualitySensor(accessory.Info{Name: ms.Name, ID: ms.ID})
			a.AirQualitySensors = append(a.AirQualitySensors, sensor)
		} else if ms.Type == "occupancy" {
			sensor := newOccupancySensor(accessory.Info{Name: ms.Name, ID: ms.ID})
			a.OccupancySensors = append(a.OccupancySensors, sensor)
		}
	}

	for _, swtch := range config.Switches {
		sw := newButton(accessory.Info{Name: swtch.Name, ID: swtch.ID})
		sw.Button.On.OnValueRemoteUpdate(sw.Callback)
		a.Switches = append(a.Switches, sw)
	}

	for _, tv := range config.Switches {
		t := newTelevision(accessory.Info{Name: tv.Name, ID: tv.ID})
		t.Tv.PowerModeSelection.OnValueRemoteUpdate(t.PowerSelect)
		a.Televisions = append(a.Televisions, t)
	}

	accs := make([]*accessory.Accessory, 0, 10)
	for _, acc := range a.ColoredLights {
		accs = append(accs, acc.Accessory)
	}
	for _, acc := range a.WhiteLights {
		accs = append(accs, acc.Accessory)
	}
	for _, acc := range a.MotionSensors {
		accs = append(accs, acc.Accessory)
	}
	for _, acc := range a.LightSensors {
		accs = append(accs, acc.Accessory)
	}
	for _, acc := range a.OccupancySensors {
		accs = append(accs, acc.Accessory)
	}
	for _, acc := range a.ContactSensors {
		accs = append(accs, acc.Accessory)
	}
	for _, acc := range a.AirQualitySensors {
		accs = append(accs, acc.Accessory)
	}
	for _, acc := range a.Switches {
		accs = append(accs, acc.Accessory)
	}

	return accs
}
