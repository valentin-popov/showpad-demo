package config

import (
	hcl "github.com/hashicorp/hcl/v2/hclsimple"
)

// API configuration structure.
type Config struct {
	Hostname string
	Port     int
	Key      string
}

type hclConf struct {
	Api *struct {
		Hostname string `hcl:"hostname"`
		Port     int    `hcl:"port"`
		Key      string `hcl:"key"`
	} `hcl:"api,block"`
}

// Load reads and parses the HCL configuration file.
func Load(filename string) (*hclConf, error) {
	cfg := &hclConf{}
	if err := hcl.DecodeFile(filename, nil, cfg); err != nil {
		return nil, err
	}

	return cfg, nil

}

// Parse validates and converts the raw HCL configuration into a Config instance.
func (rawconf *hclConf) Parse() (*Config, error) {

	if rawconf.Api.Hostname == "" {
		return nil, ErrInvalidAPIHostname
	}

	if rawconf.Api.Key == "" {
		return nil, ErrInvalidAPIKey
	}

	return &Config{

		Hostname: rawconf.Api.Hostname,
		Port:     rawconf.Api.Port,
		Key:      rawconf.Api.Key,
	}, nil

}
