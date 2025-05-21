package config

import (
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/lecture/lectureservice"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/user/userservice"
)

var (
	GlobalServerConfig ServerConfig
	GlobalEtcdConfig   EtcdConfig
	UserClient         userservice.Client
	LectureClient      lectureservice.Client
)
