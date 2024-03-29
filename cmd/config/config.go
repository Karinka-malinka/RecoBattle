package config

import (
	"github.com/BurntSushi/toml"
)

type Cnf struct {
	ApiServer ApiServer
	YandexAsr YandexAsr
}

type YandexAsr struct {
	YandexFolderId  string
	YandexKey       string
	YandexAsrUri    string
	Format          string
	SampleRateHertz string
}

type ApiServer struct {
	SecretKeyForAccessToken     string
	SecretKeyForRefreshToken    string
	AccessTokenExpiresAt        uint
	RefreshTokenExpiresAt       uint
	SecretKeyForHashingPassword string
}

type ConfigData struct {
	RunAddr         string
	ConfigASR       string
	DatabaseDSN     string
	PathFileStorage string
}

func NewConfig() *ConfigData {
	return &ConfigData{}
}

func (conf *ConfigData) GetConfig(file string) (*Cnf, error) {

	var cfg *Cnf

	_, err := toml.DecodeFile(file, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
