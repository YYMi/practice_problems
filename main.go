package main

import (
	"fmt"
	"log"
	"practice_problems/deepseek"
	"practice_problems/global"
	"practice_problems/initialize"
	"practice_problems/router"

	"github.com/spf13/viper"
)

func main() {
	// 1. 加载配置 (包括 OSS 配置)
	loadConfig()
	// 2. 初始化日志 (放在最前面)
	initialize.InitLogger()
	// 3. 初始化 SQLite
	initialize.InitSQLite()
	defer global.DB.Close() // 程序结束时关闭数据库
	deepseek.Init(global.DeepseekApiKey)
	// 4. 初始化路由
	r := router.InitRouter()

	// 5. 启动 Web 服务
	port := ":19527" // 你可以从配置文件读取端口
	fmt.Printf("服务正在启动，监听端口 %s...\n", port)

	if err := r.Run(port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// 简单的配置加载逻辑
func loadConfig() {
	v := viper.New()
	v.SetConfigFile("config.yaml")

	// 读取配置文件 (如果文件不存在也不报错)
	if err := v.ReadInConfig(); err != nil {
		// 配置文件不存在时不报错，使用默认配置
		fmt.Println("配置文件不存在，使用默认配置")
		return
	}

	// 读取 OSS 配置
	global.OssEndpoint = v.GetString("aliyun.endpoint")
	global.OssPrefix = v.GetString("aliyun.prefix")
	global.OssAccessKeyID = v.GetString("aliyun.access_key_id")
	global.OssAccessKeySecret = v.GetString("aliyun.access_key_secret")
	global.OssBucket = v.GetString("aliyun.bucket")
	global.OssInternalEndpoint = v.GetString("aliyun.internal_endpoint")
	global.DeepseekApiKey = v.GetString("deepseek.api_key")
	global.VoiceAppKey = v.GetString("aliyun.voice_app_key")

	// 如果配置了 OSS，打印日志
	if global.OssEndpoint != "" {
		fmt.Printf("OSS 已配置: %s\n", global.GetOssUrl())
		if global.IsOssUploadEnabled() {
			fmt.Println("OSS 上传已启用")
		} else {
			fmt.Println("OSS 上传未启用 (缺少 AccessKey 配置)")
		}
	}
}

func runBusinessLogic() {
	// 这里模拟你的业务代码
	// 直接使用 global.DB 进行查询
	var version string
	err := global.DB.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		log.Printf("查询出错: %v", err)
		return
	}
	fmt.Printf("当前数据库版本: %s\n", version)
}
