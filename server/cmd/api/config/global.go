package config

import (
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/agent/agentservice"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/chat/chatservice"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/lecture/lectureservice"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/push/pushservice"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz/quizservice"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/user/userservice"
)

var (
	GlobalServerConfig ServerConfig
	GlobalEtcdConfig   EtcdConfig
	UserClient         userservice.Client
	LectureClient      lectureservice.Client
	ChatClient         chatservice.Client
	PushClient         pushservice.Client
	QuizClient         quizservice.Client
	AgentClient        agentservice.Client
)
