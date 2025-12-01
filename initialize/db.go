package initialize

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"practice_problems/global" // 确保这里是你项目实际的 global 包路径
	"time"

	_ "modernc.org/sqlite" // 引入纯 Go 版 SQLite 驱动
)

// ... (上面是你原有的 InitMySQL 代码，保持不动) ...

// InitSQLite 初始化 SQLite 数据库连接
// 替代 InitMySQL，将连接赋值给 global.DB
func InitSQLite() {
	// 1. 定义数据库路径 (按你要求放在 uploads 文件夹下)
	// 建议使用相对路径或配置读取，这里演示固定路径
	dbDir := "./uploads"
	dbName := "data.db" // 数据库文件名
	dbPath := filepath.Join(dbDir, dbName)

	// 检查并创建目录
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		err := os.MkdirAll(dbDir, 0755)
		if err != nil {
			log.Fatalf("创建数据库目录失败: %v", err)
		}
	}

	// 2. 打开数据库连接
	var err error
	// 注意：这里赋值给全局变量 global.DB
	global.DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("打开 SQLite 数据库失败: %v", err)
	}

	// 3. 启用 WAL 模式 (性能优化关键)
	// 提高并发性能，防止写操作阻塞读操作
	if _, err := global.DB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Printf("警告: 启用 WAL 模式失败: %v", err)
	}

	// 4. 设置数据同步模式
	// FULL 模式确保数据安全，配合 WAL 性能依然很好
	if _, err := global.DB.Exec("PRAGMA synchronous=FULL"); err != nil {
		log.Printf("警告: 设置 synchronous 失败: %v", err)
	}

	// 5. 启用外键支持 (SQLite 默认是关闭外键约束的，建议开启)
	if _, err := global.DB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		log.Printf("警告: 启用外键约束失败: %v", err)
	}

	// 6. 设置连接池参数
	// SQLite 是文件锁，并发写入能力有限，虽然 WAL 改善了，但 OpenConns 不宜过大
	// 不过对于大多数 Web 应用，默认设置或者稍微限制一下即可
	global.DB.SetMaxIdleConns(5)
	global.DB.SetMaxOpenConns(100)
	global.DB.SetConnMaxLifetime(time.Hour)

	// 7. 测试连接
	if err := global.DB.Ping(); err != nil {
		log.Fatalf("连接 SQLite 数据库失败: %v", err)
	}

	log.Printf("✅ SQLite 连接成功！数据库文件位于: %s", dbPath)

	// 8. 执行建表操作
	// 第一次运行时，因为是一个空文件，你需要在这里创建表结构
	initSQLiteTables(global.DB)

	//MigrateDataFromMySQL()

}

// initSQLiteTables 初始化 SQLite 表结构
// 包含：建表语句 + 自动更新时间的触发器
func initSQLiteTables(db *sql.DB) {
	// 定义 SQL 语句切片，按顺序执行
	// 顺序很重要：先创建被依赖的表 (subjects)，再创建依赖别人的表 (categories...)
	sqlStmts := []string{
		// ==========================
		// 1. 表：subjects (科目)
		// ==========================
		`CREATE TABLE IF NOT EXISTS subjects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			status INTEGER DEFAULT 1, -- 1启用，0禁用
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		// 触发器：subjects 更新时自动刷新 update_time
		`CREATE TRIGGER IF NOT EXISTS trg_update_subjects_time 
		 AFTER UPDATE ON subjects 
		 BEGIN 
			UPDATE subjects SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 2. 表：knowledge_categories (知识点分类)
		// ==========================
		`CREATE TABLE IF NOT EXISTS knowledge_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			subject_id INTEGER NOT NULL,
			categorie_name TEXT NOT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			-- 外键约束
			CONSTRAINT fk_subject FOREIGN KEY (subject_id) REFERENCES subjects (id)
		);`,
		// 触发器
		`CREATE TRIGGER IF NOT EXISTS trg_update_categories_time 
		 AFTER UPDATE ON knowledge_categories 
		 BEGIN 
			UPDATE knowledge_categories SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 3. 表：knowledge_points (知识点详情)
		// ==========================
		`CREATE TABLE IF NOT EXISTS knowledge_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			categorie_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			reference_links TEXT,      -- JSON 在 SQLite 中存为 TEXT
			local_image_names TEXT,    -- JSON 在 SQLite 中存为 TEXT
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_categorie FOREIGN KEY (categorie_id) REFERENCES knowledge_categories (id)
		);`,
		// 触发器
		`CREATE TRIGGER IF NOT EXISTS trg_update_points_time 
		 AFTER UPDATE ON knowledge_points 
		 BEGIN 
			UPDATE knowledge_points SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 4. 表：questions (题目)
		// ==========================
		`CREATE TABLE IF NOT EXISTS questions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			knowledge_point_id INTEGER NOT NULL,
			question_text TEXT NOT NULL,
			option1 TEXT,
			option1_img TEXT,
			option2 TEXT,
			option2_img TEXT,
			option3 TEXT,
			option3_img TEXT,
			option4 TEXT,
			option4_img TEXT,
			correct_answer INTEGER NOT NULL, -- 1-4
			explanation TEXT,
			note TEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_point FOREIGN KEY (knowledge_point_id) REFERENCES knowledge_points (id)
		);`,
		// 触发器
		`CREATE TRIGGER IF NOT EXISTS trg_update_questions_time 
		 AFTER UPDATE ON questions 
		 BEGIN 
			UPDATE questions SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,
	}

	log.Println("⏳ 正在检查并初始化 SQLite 表结构...")

	for _, stmt := range sqlStmts {
		_, err := db.Exec(stmt)
		if err != nil {
			// 打印出错误的 SQL 方便调试
			log.Printf("❌ 执行 SQL 失败:\n%s\n错误信息: %v", stmt, err)
			// 这里可以选择是否 panic，建议开发阶段 panic，生产环境 log
			log.Panic("数据库初始化失败，程序退出")
		}
	}

	log.Println("✅ SQLite 表结构初始化完成")
}
