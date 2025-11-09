package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/gommon/log"
)

var NewVersionIsAvailable string

func NewVersionCheck() string {
	tr := &http.Transport{
		MaxIdleConns:          5,
		IdleConnTimeout:       30 * time.Second,
		DisableCompression:    true,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 3 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}
	client := &http.Client{Transport: tr}
	url := "http://www.pavogroup.top/software/goeasyjson/version.json"
	resp, err := client.Get(url)
	if err != nil {
		log.Error("New version check Error:", err)
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		var data map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			log.Info("New version Error decoding JSON:", err)
			return ""
		}
		// Extract version and details as strings
		newVersionStr, ok := data["version"].(string)
		if !ok {
			// try to handle numeric version fields by formatting
			if vnum, ok2 := data["version"].(float64); ok2 {
				newVersionStr = fmt.Sprintf("%v", vnum)
			} else {
				log.Info("New version check: Invalid version format")
				return ""
			}
		}

		upgradeDetails := ""
		if d, ok := data["details"].(string); ok {
			upgradeDetails = d
		}

		// If different from current version, set global message
		if newVersionStr != CurrentVersion {
			NewVersionIsAvailable = "A new version is available, pls run GoEasyJson -upgrade to update. \n" + "Details: " + upgradeDetails
			Lg.Info(NewVersionIsAvailable)
		}
	}
	return NewVersionIsAvailable

}
