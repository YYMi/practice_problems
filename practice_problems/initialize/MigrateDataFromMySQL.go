package initialize

import (
	"database/sql"
	"fmt"
	"log"
	"practice_problems/global" // 替换为你实际的 global 包路径

	// 👇 必须添加这一行！否则 sql.Open("mysql") 不知道去哪里找驱动
	_ "github.com/go-sql-driver/mysql"

	// 👇 这是你已经加了的 SQLite 驱动
	_ "modernc.org/sqlite"
)

// MigrateDataFromMySQL 从 MySQL 迁移数据到 SQLite
// 注意：请确保 MySQL 服务是开启的，且 global.Config.MySQL 配置正确
func MigrateDataFromMySQL() {
	log.Println("🚀 开始执行数据迁移: MySQL -> SQLite ...")

	// 1. 临时连接 MySQL (源数据库)
	// 这里我们需要手动构建 MySQL 连接，因为 global.DB 已经被 SQLite 占用了
	m := global.Config.MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.User,
		m.Password,
		m.Host,
		m.Port,
		m.DBName,
	)
	mysqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ 迁移失败: 无法连接 MySQL: %v", err)
	}
	defer mysqlDB.Close()

	if err := mysqlDB.Ping(); err != nil {
		log.Fatalf("❌ 迁移失败: MySQL 连接不通: %v", err)
	}

	// 2. 获取 SQLite 连接 (目标数据库)
	sqliteDB := global.DB
	if sqliteDB == nil {
		log.Fatal("❌ 迁移失败: SQLite 尚未初始化")
	}

	// 开启 SQLite 事务 (极大提高写入速度，保证原子性)
	tx, err := sqliteDB.Begin()
	if err != nil {
		log.Fatalf("❌ 开启事务失败: %v", err)
	}

	// 3. 定义迁移逻辑
	// 我们按依赖顺序迁移：Subjects -> Categories -> Points -> Questions

	// --- 迁移 Subjects ---
	migrateTable(mysqlDB, tx, "subjects",
		"SELECT id, name, status, create_time FROM subjects",
		"INSERT OR IGNORE INTO subjects (id, name, status, create_time) VALUES (?, ?, ?, ?)")

	// --- 迁移 Knowledge Categories ---
	migrateTable(mysqlDB, tx, "knowledge_categories",
		"SELECT id, subject_id, categorie_name, create_time FROM knowledge_categories",
		"INSERT OR IGNORE INTO knowledge_categories (id, subject_id, categorie_name, create_time) VALUES (?, ?, ?, ?)")

	// --- 迁移 Knowledge Points ---
	migrateTable(mysqlDB, tx, "knowledge_points",
		"SELECT id, categorie_id, title, content, reference_links, local_image_names, create_time FROM knowledge_points",
		"INSERT OR IGNORE INTO knowledge_points (id, categorie_id, title, content, reference_links, local_image_names, create_time) VALUES (?, ?, ?, ?, ?, ?, ?)")

	// --- 迁移 Questions ---
	migrateTable(mysqlDB, tx, "questions",
		"SELECT id, knowledge_point_id, question_text, option1, option1_img, option2, option2_img, option3, option3_img, option4, option4_img, correct_answer, explanation, note, create_time FROM questions",
		"INSERT OR IGNORE INTO questions (id, knowledge_point_id, question_text, option1, option1_img, option2, option2_img, option3, option3_img, option4, option4_img, correct_answer, explanation, note, create_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

	// 4. 提交事务
	if err := tx.Commit(); err != nil {
		log.Fatalf("❌ 事务提交失败: %v", err)
	}

	log.Println("✅ 🎉 数据迁移全部完成！MySQL 数据已成功导入 SQLite。")
}

// 通用表迁移辅助函数
func migrateTable(srcDB *sql.DB, destTx *sql.Tx, tableName string, selectSQL string, insertSQL string) {
	log.Printf("正在迁移表: %s ...", tableName)

	rows, err := srcDB.Query(selectSQL)
	if err != nil {
		log.Fatalf("查询 MySQL 表 %s 失败: %v", tableName, err)
	}
	defer rows.Close()

	count := 0
	// 动态处理列数
	cols, _ := rows.Columns()
	columnCount := len(cols)

	// 准备容器来接收数据
	values := make([]interface{}, columnCount)
	valuePtrs := make([]interface{}, columnCount)
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// 准备插入语句
	stmt, err := destTx.Prepare(insertSQL)
	if err != nil {
		log.Fatalf("准备 SQLite 插入语句失败: %v", err)
	}
	defer stmt.Close()

	for rows.Next() {
		// 1. 从 MySQL 读
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Fatalf("读取 MySQL 数据失败: %v", err)
		}

		// 2. 处理特殊类型 (MySQL 的 []byte 需要转 string，时间需要转格式)
		// SQLite 驱动通常能处理 time.Time，但 []byte (比如 text 类型) 最好手动转 string
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				values[i] = string(b)
			}
		}

		// 3. 往 SQLite 写
		if _, err := stmt.Exec(values...); err != nil {
			log.Printf("⚠️ 插入数据失败 (ID可能冲突): %v", err)
		} else {
			count++
		}
	}

	log.Printf("   -> 表 %s 迁移完成，共导入 %d 条数据", tableName, count)
}
