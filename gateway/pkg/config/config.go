package config

import (
	"fmt"
	"time"

	hcl "github.com/hashicorp/hcl/v2/hclsimple"
)

// Gateway configuration structure.
type Config struct {
	Hostname string
	Port     int
	Routes   map[string]routeConfig

	LogFile string
	DBFile  string

	UserCacheTTL time.Duration // minutes

	Api *apiConfig
}

type apiConfig struct {
	Hostname string
	Port     int
	Key      string
}

type routeConfig struct {
	Strategy string
	Capacity int
	Limit    int
}

type hclConf struct {
	Gateway *struct {
		Hostname     string `hcl:"hostname"`
		Port         int    `hcl:"port"`
		LogFile      string `hcl:"log_file"`
		DBFile       string `hcl:"db_file"`
		UserCacheTTL int    `hcl:"user_cache_ttl_minutes,optional"`
	} `hcl:"gateway,block"`

	Api *struct {
		Hostname string `hcl:"hostname"`
		Port     int    `hcl:"port"`
		Key      string `hcl:"key"`
	} `hcl:"api,block"`

	Routes []hclRoute `hcl:"routes,block"`
}

type hclRoute struct {
	Path       string  `hcl:"path"`
	Strategy   string  `hcl:"strategy"`
	Capacity   int     `hcl:"capacity,optional"`
	Limit      int     `hcl:"limit,optional"`
	WindowSize int     `hcl:"window_size,optional"`
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
	if rawconf.Gateway == nil {
		return nil, ErrInvalidAPIHostname
	}

	if rawconf.Gateway.Hostname == "" {
		return nil, ErrMissingGatewayHost
	}
	if rawconf.Gateway.Port == 0 {
		rawconf.Gateway.Port = 8080
	}

	if rawconf.Gateway.LogFile == "" {
		return nil, ErrInvalidLogFile
	}

	if rawconf.Gateway.DBFile == "" {
		return nil, ErrInvalidDBFile
	}

	if rawconf.Gateway.UserCacheTTL == 0 {
		rawconf.Gateway.UserCacheTTL = 10 // minutes
	}

	if rawconf.Api.Hostname == "" {
		return nil, ErrInvalidAPIHostname
	}
	if rawconf.Api.Port == 0 {
		rawconf.Api.Port = 8081

	}
	if rawconf.Api.Key == "" {
		return nil, ErrInvalidAPIKey
	}

	conf := &Config{
		Hostname:     rawconf.Gateway.Hostname,
		Port:         rawconf.Gateway.Port,
		LogFile:      rawconf.Gateway.LogFile,
		UserCacheTTL: time.Duration(rawconf.Gateway.UserCacheTTL),
		DBFile:       rawconf.Gateway.DBFile,

		Api: &apiConfig{
			Hostname: rawconf.Api.Hostname,
			Port:     rawconf.Api.Port,
			Key:      rawconf.Api.Key,
		},
	}

	if len(rawconf.Routes) == 0 {
		return conf, nil
	}
	routeLimits := map[string]routeConfig{}

	for _, route := range rawconf.Routes {
		if route.Path == "" {
			continue
		}

		switch route.Strategy {
		case "token_bucket":
			if route.Capacity <= 0 {
				return nil, fmt.Errorf("%w %s", ErrTokenCapacity, route.Path)
			}

			routeLimits[route.Path] = routeConfig{
				Strategy: route.Strategy,
				Capacity: route.Capacity,
			}

		case "fixed_window":
			if route.Limit <= 0 {
				return nil, fmt.Errorf("limit must be > 0 for route %s", route.Path)
			}
			if route.WindowSize == 0 {
				return nil, fmt.Errorf("window_size is required for route %s", route.Path)
			}
		default:
			return nil, fmt.Errorf("invalid strategy for route %s", route.Path)
		}
	}

	if len(routeLimits) > 0 {
		conf.Routes = routeLimits
	}

	return conf, nil
}
