package config

import (
	"os"

	hcl "github.com/hashicorp/hcl/v2/hclsimple"
)

// API configuration structure.
type Config struct {
	Address string
	Key     string
}

type hclConf struct {
	Api *struct {
		Address string `hcl:"address"`
		Key     string `hcl:"key"`
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

	if rawconf.Api.Address == "" {
		return nil, ErrInvalidAPIAddress
	}

	port := os.Getenv("PORT")
	if port != "" {
		// for google cloud
		rawconf.Api.Address = ":" + port
	}

	if rawconf.Api.Key == "" {
		return nil, ErrInvalidAPIKey
	}

	return &Config{

		Address: rawconf.Api.Address,
		Key:     rawconf.Api.Key,
	}, nil

}
