package service

import (
	"errors"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
)

func buildServer(name string, kvs kvstore.KVStore) (*config.ServerConfig, error) {

	log.Info("Building server " + name)
	serverConfig, err := config.ReadServerConfig(name, kvs)
	if err != nil {
		return nil, err
	}

	if serverConfig == nil {
		return nil, errors.New("No definition for server '" + name + "' found")
	}

	return serverConfig, nil
}
