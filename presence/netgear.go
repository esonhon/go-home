package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// netgear: Go API to retrieve devices attached to modern Netgear routers

const sessionID = "A7D88AE69687E58D9A00"

const soapActionLogin = "urn:NETGEAR-ROUTER:service:ParentalControl:1#Authenticate"
const soapLoginMessage = `\
<?xml version="1.0" encoding="utf-8" ?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
<SOAP-ENV:Header>
<SessionID xsi:type="xsd:string" xmlns:xsi="http://www.w3.org/1999/XMLSchema-instance">%s</SessionID>
</SOAP-ENV:Header>
<SOAP-ENV:Body>
<Authenticate>
  <NewUsername>%s</NewUsername>
  <NewPassword>%s</NewPassword>
</Authenticate>
</SOAP-ENV:Body>
</SOAP-ENV:Envelope>
`

const soapActionGetAttachedDevices = "urn:NETGEAR-ROUTER:service:DeviceInfo:1#GetAttachDevice"
const soapGetAttachedDevicesMesssage = `\
<?xml version="1.0" encoding="utf-8" standalone="no"?>
<SOAP-ENV:Envelope xmlns:SOAPSDK1="http://www.w3.org/2001/XMLSchema" xmlns:SOAPSDK2="http://www.w3.org/2001/XMLSchema-instance" xmlns:SOAPSDK3="http://schemas.xmlsoap.org/soap/encoding/" xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
<SOAP-ENV:Header>
<SessionID>%s</SessionID>
</SOAP-ENV:Header>
<SOAP-ENV:Body>
<M1:GetAttachDevice xmlns:M1="urn:NETGEAR-ROUTER:service:DeviceInfo:1">
</M1:GetAttachDevice>
</SOAP-ENV:Body>
</SOAP-ENV:Envelope>
`

// AttachedDevice represents a network device attached to the Netgear router via
// a wired or wireless link
type AttachedDevice struct {
	Signal   string `json:"signal"`
	IP       string `json:"ip"`
	Name     string `json:"name"`
	Mac      string `json:"mac"`
	Type     string `json:"type"`
	LinkRate string `json:"link_rate"`
}

// Netgear describes a modern Netgear router providing a SOAP interface at port
// 5000
type Netgear struct {
	host     string
	username string
	password string
	loggedIn bool
	regex    *regexp.Regexp
}

// IsLoggedIn returns true if the session has been authenticated against the
// Netgear Router or false otherwise.
func (netgear *Netgear) IsLoggedIn() bool {
	return netgear.loggedIn
}

// Login authenticates the session against the Netgear router
// On success true and nil should be returned. Otherwise false and
// the related error are returned
func (netgear *Netgear) Login() (bool, error) {
	message := fmt.Sprintf(soapLoginMessage, sessionID, netgear.username, netgear.password)

	resp, err := netgear.makeRequest(soapActionLogin, message)

	if strings.Contains(resp, "<ResponseCode>000</ResponseCode>") {
		netgear.loggedIn = true
	} else {
		netgear.loggedIn = false
	}
	return netgear.loggedIn, err
}

func (netgear *Netgear) getUrl() string {
	return fmt.Sprintf("http://%s:5000/soap/server_sa/", netgear.host)
}

func (netgear *Netgear) makeRequest(action string, message string) (string, error) {
	client := &http.Client{}

	url := netgear.getUrl()

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(message)))
	if err != nil {
		return "", err
	}
	req.Header.Add("SOAPAction", action)

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), err
}

// GetAttachedDevices queries the Netgear router for attached network
// devices and returns a list of them. If an error occures an empty list
// and the respective error is returned.
func (netgear *Netgear) GetAttachedDevices() ([]AttachedDevice, error) {
	var result []AttachedDevice

	message := fmt.Sprintf(soapGetAttachedDevicesMesssage, sessionID)
	resp, err := netgear.makeRequest(soapActionGetAttachedDevices, message)

	if strings.Contains(resp, "<ResponseCode>000</ResponseCode>") {
		re := netgear.regex.FindStringSubmatch(resp)
		if len(re) < 2 {
			err = fmt.Errorf("Invalid response code")
			return result, err
		}

		filteredDevicesStr := strings.Replace(re[1], "&lt;unknown&gt;", "unknown", -1)

		deviceStrs := strings.Split(filteredDevicesStr, "@")

		for _, deviceStr := range deviceStrs {
			fields := strings.Split(deviceStr, ";")

			signal_str := "?"
			ip_str := ""
			name_str := ""
			mac_str := ""
			type_str := "?"
			linkrate_str := "?"
			if len(fields) >= 2 {
				ip_str = fields[1]
			}
			if len(fields) >= 3 {
				name_str = fields[2]
			}
			if len(fields) >= 4 {
				mac_str = fields[3]
			}

			device := AttachedDevice{
				Signal:   signal_str,
				IP:       ip_str,
				Name:     name_str,
				Mac:      mac_str,
				Type:     type_str,
				LinkRate: linkrate_str,
			}
			result = append(result, device)
		}

	}
	return result, err
}

// NewRouter returns a new and already initialized Netgear router instance
// However, the Netgear SOAP session has not been authenticated at this point.
// Use Login() to authenticate against the router
func NewRouter(host, username, password string) *Netgear {
	router := &Netgear{
		host:     host,
		username: username,
		password: password,
		regex:    regexp.MustCompile("<NewAttachDevice>(.*)</NewAttachDevice>"),
	}
	return router
}
