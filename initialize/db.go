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
func initSQLiteTables(db *sql.DB) {
	// 定义 SQL 语句切片，按顺序执行
	// 顺序非常重要：必须先创建被 FK 引用的表 (如 users, subjects, share_codes)
	sqlStmts := []string{
		// ==========================
		// 1. 基础表：users (用户表)
		// ==========================
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,    
			user_code TEXT NOT NULL UNIQUE,   
			password TEXT NOT NULL,           
			nickname TEXT,                    
			email TEXT,                       
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_users_time 
		 AFTER UPDATE ON users BEGIN 
			UPDATE users SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 2. 基础表：subjects (科目表)
		// ==========================
		// 注意：补全了 creator_code
		`CREATE TABLE IF NOT EXISTS subjects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			status INTEGER DEFAULT 1, -- 1启用，0禁用
			creator_code TEXT,        -- 创建者代码
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_subjects_time 
		 AFTER UPDATE ON subjects BEGIN 
			UPDATE subjects SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 3. 核心业务表：share_codes (分享码主表)
		// ==========================
		// 注意：原 SQL 为 share_announcements，但为了配合 share_code_subjects 的外键，这里统一命名为 share_codes
		`CREATE TABLE IF NOT EXISTS share_codes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			creator_code TEXT NOT NULL,      -- 分享者 code
			share_code TEXT NOT NULL,        -- 分享码
			note TEXT,                       -- 备注
			status INTEGER DEFAULT 1,        -- 状态 (1:正常, 0:已删除)
			expire_time DATETIME,            -- 过期时间
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// ==========================
		// 4. 关联表：user_subjects (用户-科目绑定)
		// ==========================
		// 注意：补全了 expire_time, status, source_share_code_id
		`CREATE TABLE IF NOT EXISTS user_subjects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,      
			subject_id INTEGER NOT NULL,   
			status INTEGER DEFAULT 1,
			source_share_code_id INTEGER DEFAULT 0,
			expire_time DATETIME,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			
			CONSTRAINT uk_user_subject UNIQUE (user_id, subject_id),
			CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
			CONSTRAINT fk_subject FOREIGN KEY (subject_id) REFERENCES subjects (id) ON DELETE CASCADE
		);`,

		// ==========================
		// 5. 关联表：share_code_subjects (分享码包含的科目)
		// ==========================
		`CREATE TABLE IF NOT EXISTS share_code_subjects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			share_code_id INTEGER NOT NULL, 
			subject_id INTEGER NOT NULL,    
			
			CONSTRAINT fk_main_code FOREIGN KEY (share_code_id) REFERENCES share_codes (id) ON DELETE CASCADE,
			CONSTRAINT fk_sub_id FOREIGN KEY (subject_id) REFERENCES subjects (id) ON DELETE CASCADE
		);`,

		// ==========================
		// 6. 关联表：share_code_usage (分享码使用记录)
		// ==========================
		`CREATE TABLE IF NOT EXISTS share_code_usage (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			share_code_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			use_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			
			CONSTRAINT uk_code_user UNIQUE (share_code_id, user_id)
		);`,

		// ==========================
		// 7. 题库结构表：knowledge_categories (章节/分类)
		// ==========================
		`CREATE TABLE IF NOT EXISTS knowledge_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			subject_id INTEGER NOT NULL,
			sort_order INTEGER DEFAULT 0,
			difficulty INTEGER DEFAULT 0,
			categorie_name TEXT NOT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_subject FOREIGN KEY (subject_id) REFERENCES subjects (id) ON DELETE NO ACTION ON UPDATE NO ACTION
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_categories_time 
		 AFTER UPDATE ON knowledge_categories BEGIN 
			UPDATE knowledge_categories SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 8. 题库结构表：knowledge_points (知识点)
		// ==========================
		`CREATE TABLE IF NOT EXISTS knowledge_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			categorie_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			reference_links TEXT,
			local_image_names TEXT,
			sort_order INTEGER DEFAULT 0,
			difficulty INTEGER DEFAULT 0,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_categorie FOREIGN KEY (categorie_id) REFERENCES knowledge_categories (id) ON DELETE NO ACTION ON UPDATE NO ACTION
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_points_time 
		 AFTER UPDATE ON knowledge_points BEGIN 
			UPDATE knowledge_points SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 9. 题库结构表：questions (题目)
		// ==========================
		`CREATE TABLE IF NOT EXISTS questions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			knowledge_point_id INTEGER NOT NULL,
			question_text TEXT NOT NULL,
			option1 TEXT, option1_img TEXT,
			option2 TEXT, option2_img TEXT,
			option3 TEXT, option3_img TEXT,
			option4 TEXT, option4_img TEXT,
			correct_answer INTEGER NOT NULL, 
			explanation TEXT,
			note TEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_point FOREIGN KEY (knowledge_point_id) REFERENCES knowledge_points (id)
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_questions_time 
		 AFTER UPDATE ON questions BEGIN 
			UPDATE questions SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,
	}

	log.Println("⏳ 正在检查并初始化 SQLite 表结构...")

	for _, stmt := range sqlStmts {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Printf("❌ 执行 SQL 失败:\n%s\n错误信息: %v", stmt, err)
			// 在开发阶段，如果表结构变动大，建议删除 data.db 重新生成，
			// 或者手动处理迁移。这里 panic 提醒开发者。
			log.Panic("数据库初始化失败，程序退出")
		}
	}

	log.Println("✅ SQLite 表结构初始化完成")
}
