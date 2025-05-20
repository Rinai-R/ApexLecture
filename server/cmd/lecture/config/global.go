package config

import "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/interaction/interactionservice"

var (
	GlobalServerConfig ServerConfig
	GlobalEtcdConfig   EtcdConfig
	InteractionClient  interactionservice.Client
)
