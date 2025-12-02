package main

import (
	"fmt"
	"log"
	"practice_problems/config"
	"practice_problems/global"
	"practice_problems/initialize"
	"practice_problems/router"

	"github.com/spf13/viper"
)

func main() {
	// 1. 加载配置 (和之前一样)
	loadConfig()

	// 2. 初始化 MySQL (和之前一样)
	initialize.InitSQLite()
	defer global.DB.Close() // 程序结束时关闭数据库

	// 3. 初始化路由
	r := router.InitRouter()

	// 4. 启动 Web 服务
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
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	global.Config = &config.ServerConfig{}
	if err := v.Unmarshal(global.Config); err != nil {
		log.Fatalf("配置解析失败: %v", err)
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
