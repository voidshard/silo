package main

import (
	"github.com/voidshard/silo"
	"gopkg.in/gcfg.v1"
)

// high level config, from the point of view of the webservice
type Config struct {
	// settings specific to the HTTP server
	Server *serverSettings

	// settings intended for silo
	SiloConfig *silo.Config
}


// basic structure of the config file read in by this service
type fileConfig struct {
	Server serverSettings
	Misc miscSettings
	Store storageSettings
	Role map[string]*entity
}

// -- sections of the config file --
type serverSettings struct {
	HttpHost string
	HttpPort int
	SSLCert string
	SSLKey string
}

type storageSettings struct {
	Driver string
	Location string
}

type miscSettings struct {
	MaxDataBytes int
	MaxKeyBytes int
	EncryptionKey string
}

type entity struct {
	Id string
	Password string
	Get bool
	Put bool
	Del bool
}
// -- end sections of config file


func readConfigFile(filename string) (*fileConfig, error) {
	fcfg := &fileConfig{}
	return fcfg, gcfg.ReadFileInto(fcfg, filename)
}

// Parse config file
//
func parseConfig(filename string) (*Config, error) {
	fcfg, err := readConfigFile(filename)
	if err != nil {
		return nil, err
	}

	// convert the required bits of our http server config into a silo config
	siloConfig := silo.NewConfig()

	if fcfg.Store.Location != "" {
		siloConfig.Store.Location = fcfg.Store.Location
		siloConfig.Store.Driver = fcfg.Store.Driver
	}

	if fcfg.Misc.EncryptionKey != "" {
		siloConfig.Misc.EncryptionKey = fcfg.Misc.EncryptionKey
	}
	if fcfg.Misc.MaxKeyBytes > 0 {
		siloConfig.Misc.MaxKeyBytes = fcfg.Misc.MaxKeyBytes
	}
	if fcfg.Misc.MaxDataBytes > 0 {
		siloConfig.Misc.MaxDataBytes = fcfg.Misc.MaxDataBytes
	}

	if len(fcfg.Role) > 0 {
		susers := map[string]*silo.Role{}
		for _, u := range fcfg.Role {
			su, err := silo.NewRole(u.Id, u.Password)
			if err != nil {
				return nil, err
			}

			su.CanGet = u.Get
			su.CanPut = u.Put
			su.CanRm = u.Del

			susers[u.Id] = su
		}
		siloConfig.User = susers
	}

	return &Config{
		Server: &fcfg.Server,
		SiloConfig: siloConfig,
	}, nil
}
