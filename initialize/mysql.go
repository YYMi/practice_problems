package initialize

func InitMySQL() {
	//m := global.Config.MySQL
	//
	//// 构建 DSN (Data Source Name)
	//dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
	//	m.User,
	//	m.Password,
	//	m.Host,
	//	m.Port,
	//	m.DBName,
	//)
	//
	//var err error
	//// 将连接赋值给全局变量 global.DB
	//global.DB, err = sql.Open("mysql", dsn)
	//if err != nil {
	//	log.Fatalf("初始化数据库对象失败: %v", err)
	//}
	//
	//// 设置连接池参数
	//global.DB.SetMaxIdleConns(m.MaxIdleConns)
	//global.DB.SetMaxOpenConns(m.MaxOpenConns)
	//global.DB.SetConnMaxLifetime(time.Hour)
	//
	//// 测试连接
	//if err := global.DB.Ping(); err != nil {
	//	log.Fatalf("连接数据库失败: %v", err)
	//}
	//
	//log.Println("MySQL 连接成功！")
}
