package presence

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nanopack/mist/clients"
)

// HOME    is a state that happens when   NOT_AWAY > X seconds
// AWAY    is a state that happens when   NOT_SEEN > N seconds

type state uint32

const (
	home state = iota
	away
	leaving
	arriving
)

// LEFT    is a trigger that happens when state changes from HOME => AWAY
// ARRIVED is a trigger that happens when state changes from AWAY => HOME

// Config is the configuration needed to call presence.New
type Config struct {
	Name       string  `json:"name"`
	Host       string  `json:"host"`
	Port       int     `json:"port"`
	User       string  `json:"user"`
	Password   string  `json:"password"`
	UpdateHist int     `json:"uhist"`
	UpdateFreq float64 `json:"ufreq"`
	Devices    []struct {
		Name string `json:"name"`
		Mac  string `json:"mac"`
	} `json:"devices"`
}

func createConfig(jsondata string) (c *Config) {
	c = &Config{}
	json.Unmarshal([]byte(jsondata), c)
	return
}

// Router is the public interface with one member function to obtain devices
type provider interface {
	get(mac map[string]bool) error
}

func newProvider(name string, host string, username string, password string) provider {
	switch {
	case name == "netgear":
		return newNetgearRouter(host, username, password)
	}
	return nil
}

type Detect struct {
	Time  time.Time
	State state
}

type Member struct {
	Name    string `json:"name"`
	State   string `json:"state"`
	current Detect
	index   int
	detect  []Detect
}

type PresenceState struct {
	Time    time.Time `json:"time"`
	Members []Member  `json:"members"`
}

// Update will set the current presence state based on historical information
func (m *Member) updateCurrent(current time.Time) {
	for _, d := range m.detect {
		if d.State == home {
			m.current.Time = current
			m.current.State = home
			return
		}
	}
	m.current.Time = current
	m.current.State = away
}

func getNameOfState(state state) string {
	switch {
	case state == home:
		return "HOME"
	case state == away:
		return "AWAY"
	case state == leaving:
		return "LEAVING"
	case state == arriving:
		return "ARRIVING"
	default:
		return "HOME"
	}
}

// Home contains multiple device tracking states
type Presence struct {
	config        *Config
	provider      provider
	macToIndex    map[string]int
	macToPresence map[string]bool
	members       []*Member
	updateFreq    float64
	state         *PresenceState
}

// New will return an instance of presence.Home
func New(configjson string) *Presence {

	presence := &Presence{}
	presence.config = createConfig(configjson)
	presence.provider = newProvider(presence.config.Name, presence.config.Host, presence.config.User, presence.config.Password)
	presence.macToIndex = map[string]int{}

	updateHist := presence.config.UpdateHist
	for i, device := range presence.config.Devices {
		member := &Member{Name: device.Name}
		member.current = Detect{Time: time.Now(), State: home}
		member.index = 0
		member.detect = make([]Detect, updateHist, updateHist)
		for i := range member.detect {
			member.detect[i] = Detect{Time: time.Now(), State: away}
		}
		presence.macToIndex[device.Mac] = i
		presence.members = append(presence.members, member)
	}

	presence.updateFreq = presence.config.UpdateFreq

	return presence
}

// Presence  ...
func (p *Presence) Presence(currentTime time.Time) bool {
	result := p.provider.get(p.macToPresence)
	if result == nil {
		// All members initialize detected presence state to 'away'
		for _, m := range p.members {
			m.detect[m.index] = Detect{Time: currentTime, State: away}
			m.index = (m.index + 1) % len(m.detect)
		}
		// For any member registered at the Router mark them as 'home'
		for mac, presence := range p.macToPresence {
			mi := p.macToIndex[mac]
			m := p.members[mi]
			pi := (m.index + len(m.detect) - 1) % len(m.detect)
			if presence {
				m.detect[pi] = Detect{Time: currentTime, State: home}
			} else {
				m.detect[pi] = Detect{Time: currentTime, State: away}
			}
		}
		// Update final presence state for all members
		for _, m := range p.members {
			m.updateCurrent(currentTime)
		}

		// Build JSON structure of members
		// {"datetime":"30/12", "members": [{"name": "Faith", "state": "HOME"},{"name": "Jurgen", "state": "LEAVING"}]}
		p.state.Time = currentTime
		for _, m := range p.members {
			m.State = getNameOfState(m.current.State)
		}
		return true
	}
	return false
}

func (p *Presence) publish(client *clients.TCP) error {
	data, err := json.Marshal(p.state)
	jsonstr := string(data)
	client.Publish([]string{"presence", "state"}, jsonstr)
	return err
}

func tagsContains(tag string, tags []string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func main() {

	var presence *Presence

	for {
		client, err := clients.New("127.0.0.1:1445", "authtoken.wicked")
		if err != nil {
			fmt.Println(err)
			continue
		}

		client.Ping()
		client.Subscribe([]string{"presence"})
		client.Publish([]string{"request", "config"}, "presence")

		for {
			select {
			case msg := <-client.Messages():
				if tagsContains("config", msg.Tags) {
					presence = New(msg.Data)
				}
				break
			case <-time.After(time.Second * 10):
				if presence != nil {
					if presence.Presence(time.Now()) {
						presence.publish(client)
					}
				}
				break
			}
		}

		// Disconnect from Mist
	}
}
