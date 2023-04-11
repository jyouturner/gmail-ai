package automation

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Gmail struct {
		Credentials string `json:"credentials"`
		Token       string `json:"token"`
	} `json:"gmail"`
	RejectionCheck struct {
		URL string `json:"url"`
	} `json:"rejectionCheck"`
}

func LoadConfig(path string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
