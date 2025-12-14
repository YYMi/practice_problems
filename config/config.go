package config

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// OSSConfig 阿里云 OSS 配置
type AliyunConfig struct {
	Endpoint         string `mapstructure:"endpoint"`          // OSS Bucket 访问地址 (外网)
	Prefix           string `mapstructure:"prefix"`            // 资源路径前缀
	AccessKeyID      string `mapstructure:"access_key_id"`     // AccessKey ID
	AccessKeySecret  string `mapstructure:"access_key_secret"` // AccessKey Secret
	Bucket           string `mapstructure:"bucket"`            // Bucket 名称
	InternalEndpoint string `mapstructure:"internal_endpoint"` // 内网 Endpoint (用于上传)
	VoiceAppKey      string `mapstructure:"voice_app_key`
}

type ServerConfig struct {
	MySQL MySQLConfig  `mapstructure:"mysql"`
	OSS   AliyunConfig `mapstructure:"oss"`
}

type DeepseekConfig struct {
	ApiKey string `mapstructure:"api_key"`
}
