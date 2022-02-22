package cvinv

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"log"
	"net/http"
)

type Device struct {
	Result struct {
		Value struct {
			Key struct {
				DeviceID string `json:"deviceId"`
			} `json:"key"`
			SoftwareVersion    string    `json:"softwareVersion"`
			ModelName          string    `json:"modelName"`
			HardwareRevision   string    `json:"hardwareRevision"`
			Fqdn               string    `json:"fqdn"`
			Hostname           string    `json:"hostname"`
			DomainName         string    `json:"domainName"`
			SystemMacAddress   string    `json:"systemMacAddress"`
			BootTime           time.Time `json:"bootTime"`
			StreamingStatus    string    `json:"streamingStatus"`
			ExtendedAttributes struct {
				FeatureEnabled struct {
					Danz bool `json:"Danz"`
					Mlag bool `json:"Mlag"`
				} `json:"featureEnabled"`
			} `json:"extendedAttributes"`
		} `json:"value"`
		Time time.Time `json:"time"`
		Type string    `json:"type"`
	} `json:"result"`
}

type CvpData struct {
	Token  string
	Url    string
	Server string
}

func (c *CvpData) CvpDevices(Token, Url, Server string) map[string]string {
	var bearer = "Bearer " + c.Token

	req, err := http.NewRequest("GET", "https://"+c.Server+c.Url, nil)
	req.Header.Add("Authorization", bearer)
	req.Header.Add("Accept", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		log.Print("response error cannot connect to cvp ", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print("Cannot Marshall the response body ", err)
	}

	f := strings.Split(string(responseData), "\n")

	devs := map[string]string{}

	for _, i := range f {
		var Dev Device
		json.Unmarshal([]byte(i), &Dev)
		if Dev.Result.Value.StreamingStatus == "STREAMING_STATUS_ACTIVE" {
			//fmt.Println(Dev.Result.Value.Fqdn, Dev.Result.Value.Key.DeviceID)
			devs[Dev.Result.Value.Fqdn] = Dev.Result.Value.Key.DeviceID
		}
	}
	return devs
}

func Log(Device string) {
	log.Println(Device, " Found")
}
