package kvstore

import (
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	consulapi "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/hashicorp/consul/api"
	"net/url"
)

func consulConfigFromEnv(u *url.URL) (*consulapi.Config, error) {
	config := consulapi.DefaultConfig()
	log.Info("Setting consul address: ", u.Host)
	config.Address = u.Host
	return config, nil
}
