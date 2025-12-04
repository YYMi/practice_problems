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

// InitSQLite 初始化 SQLite 数据库连接
// 替代 InitMySQL，将连接赋值给 global.DB
func InitSQLite() {
	// 1. 定义数据库路径
	dbDir := "./uploads"
	dbName := "data.db" // 数据库文件名
	dbPath := filepath.Join(dbDir, dbName)

	// 检查并创建目录
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		err := os.MkdirAll(dbDir, 0755)
		if err != nil {
			// 这里依然使用 log.Fatalf，因为 global.Log 可能还没初始化好 (如果 InitLogger 在 InitSQLite 之后调用)
			// 如果你确定 InitLogger 先调用，可以用 global.GetLog(c).Fatalf
			log.Fatalf("❌ 创建数据库目录失败: %v", err)
		}
	}

	// 2. 打开数据库连接
	var err error
	global.DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("❌ 打开 SQLite 数据库失败: %v", err)
	}

	// 3. 启用 WAL 模式 (性能优化关键)
	if _, err := global.DB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		// 尝试使用 global.Log (如果已初始化)，否则 fallback 到 log
		if global.Log != nil {
			global.GetLog(nil).Warnf("启用 WAL 模式失败: %v", err)
		} else {
			log.Printf("⚠️ 启用 WAL 模式失败: %v", err)
		}
	}

	// 4. 设置数据同步模式
	if _, err := global.DB.Exec("PRAGMA synchronous=FULL"); err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("设置 synchronous 失败: %v", err)
		} else {
			log.Printf("⚠️ 设置 synchronous 失败: %v", err)
		}
	}

	// 5. 启用外键支持
	if _, err := global.DB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("启用外键约束失败: %v", err)
		} else {
			log.Printf("⚠️ 启用外键约束失败: %v", err)
		}
	}

	// 6. 设置连接池参数
	global.DB.SetMaxIdleConns(5)
	global.DB.SetMaxOpenConns(100)
	global.DB.SetConnMaxLifetime(time.Hour)

	// 7. 测试连接
	if err := global.DB.Ping(); err != nil {
		log.Fatalf("❌ 连接 SQLite 数据库失败 (Ping): %v", err)
	}

	// 打印成功日志
	if global.Log != nil {
		global.GetLog(nil).Infof("✅ SQLite 连接成功！数据库文件位于: %s", dbPath)
	} else {
		log.Printf("✅ SQLite 连接成功！数据库文件位于: %s", dbPath)
	}

	// 8. 执行建表操作
	initSQLiteTables(global.DB)

	// 维护  数据
	maintainingDatabaseTables(global.DB)
}
func maintainingDatabaseTables(db *sql.DB) {
	// 1. 检查 questions 表中是否存在 note 字段
	// 使用 SQLite 特有的 PRAGMA table_info 命令获取列信息
	rows, err := db.Query("PRAGMA table_info(questions)")
	if err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("检查表结构失败: %v", err)
		} else {
			log.Printf("⚠️ 检查表结构失败: %v", err)
		}
		return
	}
	defer rows.Close()

	hasNoteColumn := false
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{} // 默认值可能是 null
		var pk int

		// 扫描每一列的信息
		err = rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			continue
		}

		// 找到了 note 字段
		if name == "note" {
			hasNoteColumn = true
			break
		}
	}

	// 2. 如果存在 note 字段，则删除它
	if hasNoteColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 questions 表包含废弃字段 'note'，正在执行删除...")
		} else {
			log.Println("检测到 questions 表包含废弃字段 'note'，正在执行删除...")
		}

		// 执行删除列操作
		_, err := db.Exec("ALTER TABLE questions DROP COLUMN note")
		if err != nil {
			// 如果删除失败（可能是 SQLite 版本过低不支持 DROP COLUMN），记录错误但不中断程序
			errMsg := "删除字段失败"
			if global.Log != nil {
				global.GetLog(nil).Errorf("%s: %v (可能是SQLite版本低于3.35.0)", errMsg, err)
			} else {
				log.Printf("❌ %s: %v (可能是SQLite版本低于3.35.0)", errMsg, err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功从 questions 表中移除 'note' 字段")
			} else {
				log.Println("✅ 已成功从 questions 表中移除 'note' 字段")
			}
		}
	}
}

// initSQLiteTables 初始化 SQLite 表结构
func initSQLiteTables(db *sql.DB) {
	// 启用外键约束
	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("启用外键支持失败: %v", err)
		} else {
			log.Printf("⚠️ 启用外键支持失败: %v", err)
		}
	}

	// 定义 SQL 语句切片
	sqlStmts := []string{
		// ... (这里是你上面发给我的那一大堆 CREATE TABLE 语句，请务必保留原样，不要删减) ...
		// 为了节省篇幅，我这里省略了 SQL 字符串的内容，请把你上一条消息里的 sqlStmts 内容完整复制过来放到这里
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
		`CREATE TABLE IF NOT EXISTS subjects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			status INTEGER DEFAULT 1,
			creator_code TEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_subjects_time 
		 AFTER UPDATE ON subjects BEGIN 
			UPDATE subjects SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 3. 业务表：share_codes (分享码定义表)
		// ==========================
		`CREATE TABLE IF NOT EXISTS share_codes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			creator_id INTEGER NOT NULL,
			duration_str TEXT NOT NULL,
			expire_time DATETIME NOT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			used_count INTEGER DEFAULT 0,
			status INTEGER DEFAULT 1
		);`,

		// ==========================
		// 4. 业务表：share_announcements (分享公告/记录表)
		// ==========================
		`CREATE TABLE IF NOT EXISTS share_announcements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			creator_code TEXT NOT NULL,
			share_code TEXT NOT NULL,
			note TEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			expire_time DATETIME,
			status INTEGER DEFAULT 1
		);`,

		// ==========================
		// 5. 关联表：user_subjects (用户-科目绑定)
		// ==========================
		`CREATE TABLE IF NOT EXISTS user_subjects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			subject_id INTEGER NOT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			expire_time DATETIME,
			status INTEGER DEFAULT 1,
			source_share_code_id INTEGER DEFAULT 0,
			CONSTRAINT uk_user_subject UNIQUE (user_id, subject_id),
			CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE NO ACTION,
			CONSTRAINT fk_subject FOREIGN KEY (subject_id) REFERENCES subjects (id) ON DELETE CASCADE ON UPDATE NO ACTION
		);`,

		// ==========================
		// 6. 关联表：share_code_subjects (分享码包含的科目)
		// ==========================
		`CREATE TABLE IF NOT EXISTS share_code_subjects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			share_code_id INTEGER NOT NULL,
			subject_id INTEGER NOT NULL,
			CONSTRAINT fk_main_code FOREIGN KEY (share_code_id) REFERENCES share_codes (id) ON DELETE CASCADE ON UPDATE NO ACTION,
			CONSTRAINT fk_sub_id FOREIGN KEY (subject_id) REFERENCES subjects (id) ON DELETE CASCADE ON UPDATE NO ACTION
		);`,

		// ==========================
		// 7. 关联表：share_code_usage (分享码使用记录)
		// ==========================
		`CREATE TABLE IF NOT EXISTS share_code_usage (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			share_code_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			use_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uk_code_user UNIQUE (share_code_id, user_id)
		);`,

		// ==========================
		// 8. 题库结构表：knowledge_categories (章节/分类)
		// ==========================
		`CREATE TABLE IF NOT EXISTS knowledge_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			subject_id INTEGER NOT NULL,
			sort_order INTEGER DEFAULT 0,
			categorie_name TEXT NOT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			difficulty INTEGER DEFAULT 0,
			CONSTRAINT fk_subject FOREIGN KEY (subject_id) REFERENCES subjects (id) ON DELETE NO ACTION ON UPDATE NO ACTION
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_categories_time 
		 AFTER UPDATE ON knowledge_categories BEGIN 
			UPDATE knowledge_categories SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 9. 题库结构表：knowledge_points (知识点)
		// ==========================
		`CREATE TABLE IF NOT EXISTS knowledge_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			categorie_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			reference_links TEXT,
			local_image_names TEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			sort_order INTEGER DEFAULT 0,
			difficulty INTEGER DEFAULT 0,
			CONSTRAINT fk_categorie FOREIGN KEY (categorie_id) REFERENCES knowledge_categories (id) ON DELETE NO ACTION ON UPDATE NO ACTION
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_points_time 
		 AFTER UPDATE ON knowledge_points BEGIN 
			UPDATE knowledge_points SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 10. 题库结构表：questions (题目)
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
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_point FOREIGN KEY (knowledge_point_id) REFERENCES knowledge_points (id) ON DELETE NO ACTION ON UPDATE NO ACTION
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_questions_time 
		 AFTER UPDATE ON questions BEGIN 
			UPDATE questions SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 11. 用户题目备注表 (新增)
		// ==========================
		`CREATE TABLE IF NOT EXISTS question_user_notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			question_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			note TEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			
			-- 联合唯一索引：同一个用户对同一道题只能有一个备注
			CONSTRAINT uk_user_question UNIQUE (user_id, question_id),
			
			-- 外键约束
			CONSTRAINT fk_qun_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
			CONSTRAINT fk_qun_question FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE
		);`,

		// 触发器：更新时间自动更新
		`CREATE TRIGGER IF NOT EXISTS trg_update_question_notes_time 
		 AFTER UPDATE ON question_user_notes BEGIN 
			UPDATE question_user_notes SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,
	}

	if global.Log != nil {
		global.GetLog(nil).Info("⏳ 正在检查并初始化 SQLite 表结构...")
	} else {
		log.Println("⏳ 正在检查并初始化 SQLite 表结构...")
	}

	for _, stmt := range sqlStmts {
		_, err := db.Exec(stmt)
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Fatalf("❌ 执行 SQL 失败:\n%s\n错误信息: %v", stmt, err)
			} else {
				log.Printf("❌ 执行 SQL 失败:\n%s\n错误信息: %v", stmt, err)
				log.Panic("数据库初始化失败，程序退出")
			}
		}
	}

	if global.Log != nil {
		global.GetLog(nil).Info("✅ SQLite 表结构初始化完成")
	} else {
		log.Println("✅ SQLite 表结构初始化完成")
	}
}
