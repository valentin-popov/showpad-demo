package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	hcl "github.com/hashicorp/hcl/v2/hclsimple"
)

// Gateway configuration structure.
type Config struct {
	Address string
	Routes  map[string]routeConfig

	LogFile string
	DBFile  string

	UserCacheTTL time.Duration // minutes

	Api *apiConfig
}

type apiConfig struct {
	Address string
	Key     string
}

type routeConfig struct {
	Strategy     string
	BucketCap    int // for token bucket
	WindowLength int // for fixed window, seconds
	SqlTable     string
}

type hclConf struct {
	Gateway *struct {
		Address      string `hcl:"address"`
		LogFile      string `hcl:"log_file"`
		DBFile       string `hcl:"db_file"`
		UserCacheTTL int    `hcl:"user_cache_ttl_minutes,optional"`
	} `hcl:"gateway,block"`

	Api *struct {
		Address string `hcl:"address"`
		Key     string `hcl:"key"`
	} `hcl:"api,block"`

	Routes []hclRoute `hcl:"routes,block"`
}

type hclRoute struct {
	Path       string `hcl:"path"`
	Strategy   string `hcl:"strategy"`
	Capacity   int    `hcl:"capacity,optional"`
	Limit      int    `hcl:"limit,optional"`
	WindowSize int    `hcl:"window_size,optional"`
	SqlTable   string `hcl:"sql_table,optional"`
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
		return nil, ErrInvalidAPIAddress
	}

	port := os.Getenv("PORT")
	if port != "" {
		// for google cloud
		rawconf.Gateway.Address = ":" + port
	}

	if rawconf.Gateway.Address == "" {
		return nil, ErrMissingGatewayAddress
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

	envApiAddr := os.Getenv("API_ADDRESS")
	if envApiAddr != "" {
		// for google cloud
		rawconf.Api.Address = envApiAddr
	}
	if !strings.HasPrefix(rawconf.Api.Address, "http") {
		rawconf.Api.Address = "http://" + rawconf.Api.Address
	}

	if rawconf.Api.Address == "" {
		return nil, ErrInvalidAPIAddress
	}
	if rawconf.Api.Key == "" {
		return nil, ErrInvalidAPIKey
	}

	conf := &Config{
		Address:      rawconf.Gateway.Address,
		LogFile:      rawconf.Gateway.LogFile,
		UserCacheTTL: time.Duration(rawconf.Gateway.UserCacheTTL),
		DBFile:       rawconf.Gateway.DBFile,

		Api: &apiConfig{
			Address: rawconf.Api.Address,
			Key:     rawconf.Api.Key,
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
				Strategy:  route.Strategy,
				BucketCap: route.Capacity,
			}

		case "fixed_window":

			if route.WindowSize == 0 {
				return nil, fmt.Errorf("%w %s", ErrWindowSize, route.Path)
			}
			if route.SqlTable == "" {
				route.SqlTable = "request_count"
			}
			routeLimits[route.Path] = routeConfig{
				Strategy:     route.Strategy,
				WindowLength: route.WindowSize,
				SqlTable:     route.SqlTable,
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
