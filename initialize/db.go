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
	// =====================================================
	// 1. 检查 questions 表中是否存在 note 字段并删除
	// =====================================================
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

	// 如果存在 note 字段，则删除它
	if hasNoteColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 questions 表包含废弃字段 'note'，正在执行删除...")
		} else {
			log.Println("检测到 questions 表包含废弃字段 'note'，正在执行删除...")
		}

		// 执行删除列操作
		_, err := db.Exec("ALTER TABLE questions DROP COLUMN note")
		if err != nil {
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

	// =====================================================
	// 2. 检查 users 表中是否存在 is_admin 字段，不存在则添加
	// =====================================================
	userRows, err := db.Query("PRAGMA table_info(users)")
	if err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("检查 users 表结构失败: %v", err)
		} else {
			log.Printf("⚠️ 检查 users 表结构失败: %v", err)
		}
		return
	}
	defer userRows.Close()

	hasIsAdminColumn := false
	for userRows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{}
		var pk int

		err = userRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			continue
		}

		if name == "is_admin" {
			hasIsAdminColumn = true
			break
		}
	}

	// 如果不存在 is_admin 字段，则添加
	if !hasIsAdminColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 users 表缺少 'is_admin' 字段，正在添加...")
		} else {
			log.Println("检测到 users 表缺少 'is_admin' 字段，正在添加...")
		}

		_, err := db.Exec("ALTER TABLE users ADD COLUMN is_admin INTEGER DEFAULT 0")
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Errorf("添加 is_admin 字段失败: %v", err)
			} else {
				log.Printf("❌ 添加 is_admin 字段失败: %v", err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功向 users 表添加 'is_admin' 字段")
			} else {
				log.Println("✅ 已成功向 users 表添加 'is_admin' 字段")
			}
		}
	}

	// =====================================================
	// 3. 检查 users 表中是否存在 totp_secret 字段，不存在则添加
	// =====================================================
	totpRows, err := db.Query("PRAGMA table_info(users)")
	if err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("检查 users 表结构失败: %v", err)
		} else {
			log.Printf("⚠️ 检查 users 表结构失败: %v", err)
		}
		return
	}
	defer totpRows.Close()

	hasTotpSecretColumn := false
	for totpRows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{}
		var pk int

		err = totpRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			continue
		}

		if name == "totp_secret" {
			hasTotpSecretColumn = true
			break
		}
	}

	// 如果不存在 totp_secret 字段，则添加
	if !hasTotpSecretColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 users 表缺少 'totp_secret' 字段，正在添加...")
		} else {
			log.Println("检测到 users 表缺少 'totp_secret' 字段，正在添加...")
		}

		_, err := db.Exec("ALTER TABLE users ADD COLUMN totp_secret TEXT")
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Errorf("添加 totp_secret 字段失败: %v", err)
			} else {
				log.Printf("❌ 添加 totp_secret 字段失败: %v", err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功向 users 表添加 'totp_secret' 字段")
			} else {
				log.Println("✅ 已成功向 users 表添加 'totp_secret' 字段")
			}
		}
	}

	// =====================================================
	// 4. 检查 users 表中是否存在 status 字段，不存在则添加
	// =====================================================
	statusRows, err := db.Query("PRAGMA table_info(users)")
	if err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("检查 users 表结构失败: %v", err)
		} else {
			log.Printf("⚠️ 检查 users 表结构失败: %v", err)
		}
		return
	}
	defer statusRows.Close()

	hasStatusColumn := false
	hasLastLoginTimeColumn := false
	hasAiQuotaColumn := false
	for statusRows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{}
		var pk int

		err = statusRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			continue
		}

		if name == "status" {
			hasStatusColumn = true
		}
		if name == "last_login_time" {
			hasLastLoginTimeColumn = true
		}
		if name == "ai_quota" {
			hasAiQuotaColumn = true
		}
	}

	// 如果不存在 status 字段，则添加
	if !hasStatusColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 users 表缺少 'status' 字段，正在添加...")
		} else {
			log.Println("检测到 users 表缺少 'status' 字段，正在添加...")
		}

		_, err := db.Exec("ALTER TABLE users ADD COLUMN status INTEGER DEFAULT 0")
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Errorf("添加 status 字段失败: %v", err)
			} else {
				log.Printf("❌ 添加 status 字段失败: %v", err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功向 users 表添加 'status' 字段")
			} else {
				log.Println("✅ 已成功向 users 表添加 'status' 字段")
			}
		}
	}

	// 如果不存在 last_login_time 字段，则添加
	if !hasLastLoginTimeColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 users 表缺少 'last_login_time' 字段，正在添加...")
		} else {
			log.Println("检测到 users 表缺少 'last_login_time' 字段，正在添加...")
		}

		_, err := db.Exec("ALTER TABLE users ADD COLUMN last_login_time DATETIME")
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Errorf("添加 last_login_time 字段失败: %v", err)
			} else {
				log.Printf("❌ 添加 last_login_time 字段失败: %v", err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功向 users 表添加 'last_login_time' 字段")
			} else {
				log.Println("✅ 已成功向 users 表添加 'last_login_time' 字段")
			}
		}
	}

	// 如果不存在 ai_quota 字段，则添加 (AI面试时长，单位秒，默认0)
	if !hasAiQuotaColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 users 表缺少 'ai_quota' 字段，正在添加...")
		} else {
			log.Println("检测到 users 表缺少 'ai_quota' 字段，正在添加...")
		}

		_, err := db.Exec("ALTER TABLE users ADD COLUMN ai_quota INTEGER DEFAULT 0")
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Errorf("添加 ai_quota 字段失败: %v", err)
			} else {
				log.Printf("❌ 添加 ai_quota 字段失败: %v", err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功向 users 表添加 'ai_quota' 字段")
			} else {
				log.Println("✅ 已成功向 users 表添加 'ai_quota' 字段")
			}
		}
	}

	// =====================================================
	// 4.1 检查 users 表中是否存在 login_ips 字段，不存在则添加
	// =====================================================
	loginIpsRows, err := db.Query("PRAGMA table_info(users)")
	if err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("检查 users 表 login_ips 字段失败: %v", err)
		} else {
			log.Printf("⚠️ 检查 users 表 login_ips 字段失败: %v", err)
		}
	} else {
		defer loginIpsRows.Close()

		hasLoginIpsColumn := false
		for loginIpsRows.Next() {
			var cid int
			var name string
			var ctype string
			var notnull int
			var dfltValue interface{}
			var pk int

			err = loginIpsRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
			if err != nil {
				continue
			}

			if name == "login_ips" {
				hasLoginIpsColumn = true
				break
			}
		}

		// 如果不存在 login_ips 字段，则添加
		if !hasLoginIpsColumn {
			if global.Log != nil {
				global.GetLog(nil).Info("检测到 users 表缺少 'login_ips' 字段，正在添加...")
			} else {
				log.Println("检测到 users 表缺少 'login_ips' 字段，正在添加...")
			}

			// 使用 JSON 数组存储多个IP记录：[{"ip":"xxx","time":"xxx"}, ...]
			_, err := db.Exec("ALTER TABLE users ADD COLUMN login_ips TEXT DEFAULT '[]'")
			if err != nil {
				if global.Log != nil {
					global.GetLog(nil).Errorf("添加 login_ips 字段失败: %v", err)
				} else {
					log.Printf("❌ 添加 login_ips 字段失败: %v", err)
				}
			} else {
				if global.Log != nil {
					global.GetLog(nil).Info("✅ 已成功向 users 表添加 'login_ips' 字段")
				} else {
					log.Println("✅ 已成功向 users 表添加 'login_ips' 字段")
				}
			}
		}
	}

	// =====================================================
	// 5. 检查 knowledge_points 表中是否存在 video_url 字段，不存在则添加
	// =====================================================
	kpRows, err := db.Query("PRAGMA table_info(knowledge_points)")
	if err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("检查 knowledge_points 表结构失败: %v", err)
		} else {
			log.Printf("⚠️ 检查 knowledge_points 表结构失败: %v", err)
		}
		return
	}
	defer kpRows.Close()

	hasVideoUrlColumn := false
	for kpRows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{}
		var pk int

		err = kpRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			continue
		}

		if name == "video_url" {
			hasVideoUrlColumn = true
			break
		}
	}

	// 如果不存在 video_url 字段，则添加
	if !hasVideoUrlColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 knowledge_points 表缺少 'video_url' 字段，正在添加...")
		} else {
			log.Println("检测到 knowledge_points 表缺少 'video_url' 字段，正在添加...")
		}

		// 注意：SQLite 中使用 TEXT 类型存储 JSON 字符串，并默认给一个空数组 '[]'
		_, err := db.Exec("ALTER TABLE knowledge_points ADD COLUMN video_url TEXT DEFAULT '[]'")
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Errorf("添加 video_url 字段失败: %v", err)
			} else {
				log.Printf("❌ 添加 video_url 字段失败: %v", err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功向 knowledge_points 表添加 'video_url' 字段")
			} else {
				log.Println("✅ 已成功向 knowledge_points 表添加 'video_url' 字段")
			}
		}
	}

	// =====================================================
	// 6. 检查 collection_items 表是否有 point_id、subject_id、category_id 字段，没有则添加
	// =====================================================
	collectionItemsRows, err := db.Query("PRAGMA table_info(collection_items)")
	if err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("检查 collection_items 表结构失败: %v", err)
		} else {
			log.Printf("⚠️ 检查 collection_items 表结构失败: %v", err)
		}
		return
	}
	defer collectionItemsRows.Close()

	hasPointIdColumn := false
	hasSubjectIdColumn := false
	hasCategoryIdColumn := false
	for collectionItemsRows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{}
		var pk int

		err = collectionItemsRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			continue
		}

		if name == "point_id" {
			hasPointIdColumn = true
		} else if name == "subject_id" {
			hasSubjectIdColumn = true
		} else if name == "category_id" {
			hasCategoryIdColumn = true
		}
	}

	// 如果不存在 point_id 字段，则添加
	if !hasPointIdColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 collection_items 表缺少 'point_id' 字段，正在添加...")
		} else {
			log.Println("检测到 collection_items 表缺少 'point_id' 字段，正在添加...")
		}

		_, err := db.Exec("ALTER TABLE collection_items ADD COLUMN point_id INTEGER")
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Errorf("添加 point_id 字段失败: %v", err)
			} else {
				log.Printf("❌ 添加 point_id 字段失败: %v", err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功向 collection_items 表添加 'point_id' 字段")
			} else {
				log.Println("✅ 已成功向 collection_items 表添加 'point_id' 字段")
			}
		}
	}

	// 如果不存在 subject_id 字段，则添加
	if !hasSubjectIdColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 collection_items 表缺少 'subject_id' 字段，正在添加...")
		} else {
			log.Println("检测到 collection_items 表缺少 'subject_id' 字段，正在添加...")
		}

		_, err := db.Exec("ALTER TABLE collection_items ADD COLUMN subject_id INTEGER")
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Errorf("添加 subject_id 字段失败: %v", err)
			} else {
				log.Printf("❌ 添加 subject_id 字段失败: %v", err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功向 collection_items 表添加 'subject_id' 字段")
			} else {
				log.Println("✅ 已成功向 collection_items 表添加 'subject_id' 字段")
			}
		}
	}

	// 如果不存在 category_id 字段，则添加
	if !hasCategoryIdColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 collection_items 表缺少 'category_id' 字段，正在添加...")
		} else {
			log.Println("检测到 collection_items 表缺少 'category_id' 字段，正在添加...")
		}

		_, err := db.Exec("ALTER TABLE collection_items ADD COLUMN category_id INTEGER")
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Errorf("添加 category_id 字段失败: %v", err)
			} else {
				log.Printf("❌ 添加 category_id 字段失败: %v", err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功向 collection_items 表添加 'category_id' 字段")
			} else {
				log.Println("✅ 已成功向 collection_items 表添加 'category_id' 字段")
			}
		}
	}

	// =====================================================
	// 7. 检查 collection_items 表是否有 sort_order 字段，没有则添加
	// =====================================================
	collectionItemsRows2, err := db.Query("PRAGMA table_info(collection_items)")
	if err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("检查 collection_items 表 sort_order 字段失败: %v", err)
		} else {
			log.Printf("⚠️ 检查 collection_items 表 sort_order 字段失败: %v", err)
		}
		return
	}
	defer collectionItemsRows2.Close()

	hasSortOrderColumn := false
	for collectionItemsRows2.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{}
		var pk int

		err = collectionItemsRows2.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			continue
		}

		if name == "sort_order" {
			hasSortOrderColumn = true
			break
		}
	}

	// 如果不存在 sort_order 字段，则添加
	if !hasSortOrderColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 collection_items 表缺少 'sort_order' 字段，正在添加...")
		} else {
			log.Println("检测到 collection_items 表缺少 'sort_order' 字段，正在添加...")
		}

		_, err := db.Exec("ALTER TABLE collection_items ADD COLUMN sort_order INTEGER DEFAULT 0")
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Errorf("添加 sort_order 字段失败: %v", err)
			} else {
				log.Printf("❌ 添加 sort_order 字段失败: %v", err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功向 collection_items 表添加 'sort_order' 字段")
			} else {
				log.Println("✅ 已成功向 collection_items 表添加 'sort_order' 字段")
			}
		}
	}

	// =====================================================
	// 8. 检查 collections 表是否有 is_public 字段，没有则添加
	// =====================================================
	collectionsRows, err := db.Query("PRAGMA table_info(collections)")
	if err != nil {
		if global.Log != nil {
			global.GetLog(nil).Warnf("检查 collections 表 is_public 字段失败: %v", err)
		} else {
			log.Printf("⚠️ 检查 collections 表 is_public 字段失败: %v", err)
		}
		return
	}
	defer collectionsRows.Close()

	hasIsPublicColumn := false
	for collectionsRows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{}
		var pk int

		err = collectionsRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			continue
		}

		if name == "is_public" {
			hasIsPublicColumn = true
			break
		}
	}

	// 如果不存在 is_public 字段，则添加
	if !hasIsPublicColumn {
		if global.Log != nil {
			global.GetLog(nil).Info("检测到 collections 表缺少 'is_public' 字段，正在添加...")
		} else {
			log.Println("检测到 collections 表缺少 'is_public' 字段，正在添加...")
		}

		_, err := db.Exec("ALTER TABLE collections ADD COLUMN is_public INTEGER DEFAULT 0")
		if err != nil {
			if global.Log != nil {
				global.GetLog(nil).Errorf("添加 is_public 字段失败: %v", err)
			} else {
				log.Printf("❌ 添加 is_public 字段失败: %v", err)
			}
		} else {
			if global.Log != nil {
				global.GetLog(nil).Info("✅ 已成功向 collections 表添加 'is_public' 字段")
			} else {
				log.Println("✅ 已成功向 collections 表添加 'is_public' 字段")
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
			is_admin INTEGER DEFAULT 0,
			status INTEGER DEFAULT 0,
			last_login_time DATETIME,
			login_ips TEXT DEFAULT '[]',
			totp_secret TEXT,
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
			video_url TEXT DEFAULT '[]', -- 确保初始化时就有这个字段
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
		// 11. 用户题目备注表
		// ==========================
		`CREATE TABLE IF NOT EXISTS question_user_notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			question_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			note TEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uk_user_question UNIQUE (user_id, question_id),
			CONSTRAINT fk_qun_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
			CONSTRAINT fk_qun_question FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_question_notes_time 
		 AFTER UPDATE ON question_user_notes BEGIN 
			UPDATE question_user_notes SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 12. 数据库表备注表
		// ==========================
		`CREATE TABLE IF NOT EXISTS table_comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			table_name TEXT NOT NULL UNIQUE,
			comment TEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_table_comments_time 
		 AFTER UPDATE ON table_comments BEGIN 
			UPDATE table_comments SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 13. 数据库字段备注表
		// ==========================
		`CREATE TABLE IF NOT EXISTS column_comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			table_name TEXT NOT NULL,
			column_name TEXT NOT NULL,
			comment TEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uk_table_column UNIQUE (table_name, column_name)
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_column_comments_time 
		 AFTER UPDATE ON column_comments BEGIN 
			UPDATE column_comments SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 14. 字段排序表
		// ==========================
		`CREATE TABLE IF NOT EXISTS column_orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			table_name TEXT NOT NULL,
			column_name TEXT NOT NULL,
			sort_order INTEGER DEFAULT 0,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uk_column_order UNIQUE (table_name, column_name)
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_column_orders_time 
		 AFTER UPDATE ON column_orders BEGIN 
			UPDATE column_orders SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 15. 知识点绑定表
		// ==========================
		`CREATE TABLE IF NOT EXISTS point_bindings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_subject_id INTEGER NOT NULL,
			source_point_id INTEGER NOT NULL,
			target_subject_id INTEGER NOT NULL,
			target_point_id INTEGER NOT NULL,
			bind_text TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (source_subject_id) REFERENCES subjects(id),
			FOREIGN KEY (source_point_id) REFERENCES knowledge_points(id),
			FOREIGN KEY (target_subject_id) REFERENCES subjects(id),
			FOREIGN KEY (target_point_id) REFERENCES knowledge_points(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,

		// ==========================
		// 16. 集合表
		// ==========================
		`CREATE TABLE IF NOT EXISTS collections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			is_public INTEGER DEFAULT 0,  -- 0=私有 1=公有
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_collections_time 
		 AFTER UPDATE ON collections BEGIN 
			UPDATE collections SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 17. 集合授权表
		// ==========================
		`CREATE TABLE IF NOT EXISTS collection_permissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			collection_id INTEGER NOT NULL,
			user_code TEXT NOT NULL,
			expire_time DATETIME,  -- NULL表示永久有效
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uk_collection_user UNIQUE (collection_id, user_code),
			FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_collection_permissions_time 
		 AFTER UPDATE ON collection_permissions BEGIN 
			UPDATE collection_permissions SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
		 END;`,

		// ==========================
		// 18. 集合项表（集合中的知识点）
		// ==========================
		`CREATE TABLE IF NOT EXISTS collection_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			collection_id INTEGER NOT NULL,
			point_id INTEGER NOT NULL,
			subject_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			sort_order INTEGER DEFAULT 0,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uk_collection_point UNIQUE (collection_id, point_id),
			FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE,
			FOREIGN KEY (point_id) REFERENCES knowledge_points(id) ON DELETE CASCADE,
			FOREIGN KEY (subject_id) REFERENCES subjects(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES knowledge_categories(id) ON DELETE CASCADE
		);`,

		// ==========================
		// 19. 知识点笔记表
		// ==========================
		`CREATE TABLE IF NOT EXISTS point_user_notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			point_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			note TEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uk_user_point_note UNIQUE (user_id, point_id),
			CONSTRAINT fk_pun_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
			CONSTRAINT fk_pun_point FOREIGN KEY (point_id) REFERENCES knowledge_points (id) ON DELETE CASCADE
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_update_point_notes_time 
		 AFTER UPDATE ON point_user_notes BEGIN 
			UPDATE point_user_notes SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id; 
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
