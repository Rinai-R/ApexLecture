package config

import "github.com/Rinai-R/ApexLecture/server/shared/consts"

type ServerConfig struct {
	Name     string                `json:"name"`
	Host     string                `json:"host"`
	Port     string                `json:"port"`
	Services [consts.SrvLen]string `json:"services"`
}

type EtcdConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Key  string `json:"key"`
}
