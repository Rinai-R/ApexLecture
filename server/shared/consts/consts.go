package consts

// ServerConfig
const (
	Host            = "0.0.0.0"
	ApiPort         = "10000"
	UserPort        = "10001"
	LecturePort     = "10002"
	InteractionPort = "10003"

	ApiConfig         = "./server/cmd/api/config.yaml"
	UserConfig        = "./server/cmd/user/config.yaml"
	LectureConfig     = "./server/cmd/lecture/config.yaml"
	InteractionConfig = "./server/cmd/interaction/config.yaml"

	ApiSrvPrefix         = "api"
	UserSrvPrefix        = "user"
	LectureSrvPrefix     = "lecture"
	InteractionSrvPrefix = "interaction"

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
)
