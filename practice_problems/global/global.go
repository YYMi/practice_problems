package global

import (
	"database/sql"
	"practice_problems/config"
)

var (
	DB     *sql.DB              // 全局数据库连接池对象
	Config *config.ServerConfig // 全局配置对象
)
