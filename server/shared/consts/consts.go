package consts

// ServerConfig
const (
	ApiHost     = "0.0.0.0"
	UserHost    = "0.0.0.0"
	LectureHost = "0.0.0.0"
	ChatHost    = "0.0.0.0"
	PushHost    = "0.0.0.0"
	QuizHost    = "0.0.0.0"
	AgentHost   = "0.0.0.0"

	ApiPort     = "10000"
	UserPort    = "10001"
	LecturePort = "10002"
	ChatPort    = "10003"
	PushPort    = "10004"
	QuizPort    = "10005"
	AgentPort   = "10006"

	ApiConfig     = "./server/cmd/api/config.yaml"
	UserConfig    = "./server/cmd/user/config.yaml"
	LectureConfig = "./server/cmd/lecture/config.yaml"
	ChatConfig    = "./server/cmd/chat/config.yaml"
	PushConfig    = "./server/cmd/push/config.yaml"
	QuizConfig    = "./server/cmd/quiz/config.yaml"
	AgentConfig   = "./server/cmd/agent/config.yaml"

	ApiSrvPrefix     = "api"
	UserSrvPrefix    = "user"
	LectureSrvPrefix = "lecture"
	ChatSrvPrefix    = "chat"
	PushSrvPrefix    = "push"
	QuizSrvPrefix    = "quiz"
	AgentSrvPrefix   = "agent"

	OtelEndpoint = "localhost:4317"
)

// MysqlConfig
const (
	MysqlDNS      = "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai"
	MysqlHost     = "localhost"
	MysqlUser     = "root"
	MysqlPassword = "123456"
	MysqlDatabase = "apex_db"
	MysqlPort     = "3306"
)

// RedisConfig
const (
	RedisHost     = "localhost"
	RedisPort     = "6379"
	RedisPassword = "123456"
	RedisDatabase = 0
)

// MinioConfig
const (
	MinioEndpoint  = "localhost:9000"
	MinioAccessKey = "QIEZXMXgmp537hF4oUni"
	MinioSecretKey = "Aw8FvzDjIWSqi3tkYdOTHXSFCTWV6ed3UXZ4ssPu"
	MinioBucket    = "lecture"
	MinioSecure    = false
)

// KafkaConfig
const (
	KafkaUsername     = "root"
	KafkaPassword     = "123456"
	KafkaBroker1      = "localhost:9094"
	KafkaBroker2      = "localhost:9095"
	KafkaBroker3      = "localhost:9096"
	LectureKafkaTopic = "lecture"
	LectureKafkaGroup = "lecture"
	ChatKafkaTopic    = "chat"
	ChatKafkaGroup    = "chat"
	QuizKafkaTopic    = "quiz"
	QuizKafkaGroup    = "quiz"
	AgentKafkaTopic   = "agent"
	AgentKafkaGroup   = "agent"
)

// ChatModel
const (
	AgentBaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	AgentRegion  = "cn-beijing"
	AgentAPIKey  = "a02b51ca-cf7e-4cb2-bad4-266cb3714137"
	AgentModel   = "deepseek-r1-250120"
)

// key
const (
	PrivateKey = "./server/cmd/user/keys/private_key.pem"
	PublicKey  = "./server/cmd/user/keys/public_key.pem"
)

// Other
const (
	HlogFilePath    = "./tmp/hlog/"
	KlogFilePath    = "./tmp/klog/"
	IvfFilePath     = "./tmp/record/%d/video.ivf"
	OggFilePath     = "./tmp/record/%d/audio.ogg"
	MinioObjectName = "%d:%s"

	EtcdSnowFlakeNode         = 1
	UserIDSnowFlakeNode       = 2
	LectureIDSnowFlakeNode    = 3
	AttendanceIDSnowFlakeNode = 4
	MessageIDSnowFlakeNode    = 5
)

// RedisMessage
const (
	DeleteSignal          = "__DELETE__"
	UnKnownSignal         = "__UNKNOWN__"
	RoomKey               = "lecture:room:%d"
	QuestionAnswerKey     = "quiz:answer:%d"
	WrongAnswerRecordKey  = "quiz:wrong_answer:%d"
	AcceptAnswerRecordKey = "quiz:accept_answer:%d"
	QuizLockKey           = "quiz:lock:%d"
	AudienceKey           = "lecture:audiences:%d"
	LatestMsgListKey      = "lecture:latest_msg_list:%d"
	HistoryMsgKey         = "agent:history_msg:%d:%d"
	SummaryStartedLock    = "agent:summary_started_lock:%d"
)

// some status
// agent 服务里面，有几个状态，不好梳理
// 在这里用常量表示
const (
	NotCreate  = 0
	NoSummary  = 1
	Summarized = 2
	OtherError = 3
)

// GoogleCredentials
const (
	GoogleCredentialsFile = "./path/to/your/credentials.json"
)
