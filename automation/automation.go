package main

// All automation logic is in this package
// Here we react to:
// - presence (people arriving/leaving)
// - switches (pressed)
// - events (timeofday, calendar)
// - time-based logic (morning 6:20 turn on bedroom lights)

import (
	"time"

	"github.com/jurgen-kluft/go-home/config"
	logpkg "github.com/jurgen-kluft/go-home/logging"
	"github.com/jurgen-kluft/go-home/pubsub"
)

func main() {
	auto := New()

	logger := logpkg.New("automation")
	logger.AddEntry("emitter")
	logger.AddEntry("automation")

	for {
		connected := true
		for connected {
			client := pubsub.New(config.EmitterSecrets["host"])
			register := []string{"config/automation/", "state/#/"}
			subscribe := []string{"config/automation/", "state/#/"}
			err := client.Connect("automation", register, subscribe)
			if err == nil {
				logger.LogInfo("emitter", "connected")
				for connected {
					select {
					case msg := <-client.InMsgs:
						topic := msg.Topic()
						if topic == "automation/config/" {
						} else if topic == "client/disconnected/" {
							connected = false
							logger.LogInfo("emitter", "disconnected")
						}
					case <-time.After(time.Second * 30):
						auto.HandleTime(time.Now())
					}
				}
			}
			if err != nil {
				logger.LogError("automation", err.Error())
			}
			time.Sleep(1 * time.Second)
		}
		// Wait for 5 seconds before retrying
		time.Sleep(5 * time.Second)
	}
}

type Automation struct {
	sensors             map[string]string
	presence            map[string]bool
	timeofday           string
	lastmotion          time.Time
	lastmotionFrontDoor time.Time
}

func New() *Automation {
	auto := &Automation{}
	return auto
}

func (a *Automation) FamilyIsHome() bool {
	// This can be a lot more complex:

	// Motion sensors will mark if people are home and this will be reset when the front-door openened.
	// When presence shows no devices on the network but there is still motion detected we still should
	// regard that family is at home.
	if len(a.presence) == 0 {
		if time.Now().Sub(a.lastmotion) > (time.Minute * 12) {
			return false
		}
	}

	return len(a.presence) > 0
}

func (a *Automation) IsSensor(name string, value string) bool {
	v, e := a.sensors[name]
	return e && (v == value)
}

func (a *Automation) TurnOnLight(name string) {

}
func (a *Automation) TurnOffLight(name string) {

}
func (a *Automation) ToggleLight(name string) {

}
func (a *Automation) TurnOffSwitch(name string) {

}
func (a *Automation) ToggleSwitch(name string) {

}
func (a *Automation) TurnOffTV(name string) {

}

func (a *Automation) HandleEvent(domain string, product string, name string, valuetype string, value string) {
	if domain == "sensor" && product == "calendar" && name == "tod" {
		a.HandleTimeOfDay(value)
	}
}

func (a *Automation) HandleTimeOfDay(to string) {
	a.timeofday = to
	switch to {
	case "morning":
		a.TurnOffLight("Kitchen")
		a.TurnOffLight("Living Room")
	case "lunch":
		if a.FamilyIsHome() {
			a.TurnOnLight("Kitchen")
		}
	case "bedtime":
		if a.FamilyIsHome() {
			if a.IsSensor("sensor.calendar.jennifer", "school") {
				a.TurnOnLight("Jennifer")
			}
			if a.IsSensor("sensor.calendar.sophia", "school") {
				a.TurnOnLight("Sophia")
			}
		}
	case "sleeptime":
		if a.FamilyIsHome() {
			a.TurnOnLight("Bedroom")
		}
	case "night":
		if a.IsSensor("sensor.calendar.jennifer", "school") {
			a.TurnOffLight("Bedroom")
		}
		a.TurnOffLight("Kitchen")
		a.TurnOffLight("Living Room")
		a.TurnOffLight("Jennifer")
		a.TurnOffLight("Sophia")
		a.TurnOffLight("Front door hall light")
	}
}

func (a *Automation) HandleSensor(product string, name string, valuetype string, value string) {
	if product == "xiaomi" && name == "motion_sensor_158d0001a9113b" {
		if value == "on" {
			a.lastmotion = time.Now() // Update the time we last detected motion
			a.lastmotionFrontDoor = time.Now()

			if a.timeofday == "breakfast" {
				a.TurnOnLight("Kitchen")
				a.TurnOnLight("Living Room")
			}
		} else {
			if time.Now().Sub(a.lastmotionFrontDoor) > time.Minute*5 {
				a.TurnOffLight("Front door hall light")
			}
		}
	}
	if product == "xiaomi" && name == "magnet_158d0001a9113b" {
		if value == "open" {
			a.TurnOnLight("Front door hall light")
		}
	}
}

func (a *Automation) HandleState(product string, name string, valuetype string, value string) {
	if product == "xiaomi" && name == "switch_158d00015db32c" {
		if value == "double click" {
			a.ToggleLight("Bedroom")
		}
		if value == "single click" {
			a.ToggleSwitch("wall_switch_left_158d00016da5f5")
		}
		if value == "press release" {
			a.ToggleSwitch("plug_158d00017ca3f2")
		}
	} else if product == "xiaomi" && name == "switch_158d00018dc863" {
		if value == "single click" {
			a.ToggleLight("Sophia")
		}
	}
}

func (a *Automation) HandlePresence(name string, value string) {
	if value == "away" {
		delete(a.presence, name)
		leaving := (len(a.presence) == 0)
		if leaving {
			a.HandlePresenceLeaving()
		}
	} else {
		arriving := (len(a.presence) == 0)
		a.presence[name] = true
		if arriving {
			a.HandlePresenceArriving()
		}
	}
}
func (a *Automation) HandleSwitch(name string, value string) {

}

func (a *Automation) HandlePresenceLeaving() {
	// Turn off everything
	a.TurnOffEverything()
}

func (a *Automation) HandlePresenceArriving() {
	// Depending on time-of-day
	// Turn on Kitchen
	// Turn on Living-Room

}

func (a *Automation) TurnOffEverything() {
	a.TurnOffLight("Kitchen")
	a.TurnOffLight("Living Room")
	a.TurnOffLight("Bedroom")
	a.TurnOffLight("Jennifer")
	a.TurnOffLight("Sophia")
	a.TurnOffLight("Front door hall light")

	a.TurnOffSwitch("Bedroom power plug")
	a.TurnOffSwitch("Bedroom chandelier")
	a.TurnOffSwitch("Bedroom ceiling")

	a.TurnOffTV("Samsung bedroom")
	a.TurnOffTV("Sony livingroom")
}

func (a *Automation) HandleTime(now time.Time) {
	if a.IsSensor("sensor.calendar.jennifer", "school") {
		if now.Hour() == 6 && now.Minute() == 20 {
			a.TurnOnLight("Bedroom")
		}
		if now.Hour() == 6 && now.Minute() == 30 {
			a.TurnOnLight("Jennifer")
		}
	}
	if a.IsSensor("sensor.calendar.sophia", "school") {
		if now.Hour() == 7 && now.Minute() == 10 {
			a.TurnOnLight("Bedroom")
		}
		if now.Hour() == 7 && now.Minute() == 20 {
			a.TurnOnLight("Sophia")
		}
	}
	if a.IsSensor("sensor.calendar.parents", "work") {
		if !a.IsSensor("sensor.calendar.jennifer", "school") && !a.IsSensor("sensor.calendar.sophia", "school") {
			if now.Hour() == 7 && now.Minute() == 45 {
				a.TurnOnLight("Bedroom")
			}
		}
	}

}
