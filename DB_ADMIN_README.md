# 数据库管理系统使用说明

## 功能概述

已为您的 SQLite 项目添加完整的数据库管理功能，支持可视化管理所有数据表。

## 核心特性

### 1. 用户角色管理

- 新增 `is_admin` 字段区分管理员和普通用户
- 登录接口自动返回用户角色信息
- 只有管理员可以访问数据库管理界面

### 2. 安全验证

- **Google reCAPTCHA 验证**：所有增删改操作必须通过验证码
- **JWT 鉴权**：必须登录才能访问
- **管理员权限验证**：非管理员无法访问管理接口

### 3. 数据库管理功能

- ✅ 查看所有表列表及记录数
- ✅ 查看表结构（字段、类型、约束）
- ✅ 自定义字段查询（支持字段勾选）
- ✅ 分页查询（默认 20 条/页）
- ✅ 条件查询（支持多字段组合）
- ✅ 新增数据
- ✅ 编辑数据
- ✅ 删除数据（单条/批量）
- ✅ 批量更新

## 后端接口清单

### 查询接口（不需要 reCAPTCHA）

- `GET /api/v1/admin/db/tables` - 获取所有表列表
- `GET /api/v1/admin/db/tables/:table/structure` - 获取表结构
- `GET /api/v1/admin/db/tables/:table/data` - 获取表数据（支持分页、字段选择、条件查询）

### 修改接口（需要 reCAPTCHA 验证）

- `POST /api/v1/admin/db/tables/:table/insert` - 插入数据
- `PUT /api/v1/admin/db/tables/:table/update` - 更新数据
- `DELETE /api/v1/admin/db/tables/:table/delete` - 删除数据
- `PUT /api/v1/admin/db/tables/:table/batch-update` - 批量更新
- `DELETE /api/v1/admin/db/tables/:table/batch-delete` - 批量删除

## 前端访问

访问路径：`/db-admin`

### 入口

- 主页右上角"数据库管理"按钮（仅管理员可见）
- 直接访问 `http://your-domain/db-admin`（非管理员会被拦截）

### 使用流程

1. **选择表**：左侧列表选择要操作的表
2. **选择字段**：勾选需要查询的字段（默认全选）
3. **条件查询**：展开"高级查询"，输入查询条件
4. **查询数据**：点击"查询"按钮
5. **操作数据**：
   - 新增：点击"新增"按钮
   - 编辑：点击行的"编辑"按钮
   - 删除：点击行的"删除"按钮或勾选多行后点击"批量删除"

## 设置管理员

### 方法 1：数据库直接修改

```sql
UPDATE users SET is_admin = 1 WHERE username = 'admin';
```

### 方法 2：程序修改

在 `api/user.go` 的 `CreateUser` 函数中，为第一个注册的用户设置管理员权限：

```go
// 如果是第一个用户，设置为管理员
var userCount int
global.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
isAdmin := 0
if userCount == 0 {
    isAdmin = 1
}

_, err = global.DB.Exec(
    "INSERT INTO users (username, password, user_code, nickname, email, is_admin) VALUES (?, ?, ?, ?, ?, ?)",
    req.Username, string(hash), userCode, req.Nickname, req.Email, isAdmin,
)
```

## Google reCAPTCHA 配置

### 1. 获取密钥

访问 https://www.google.com/recaptcha/admin 创建站点，选择 reCAPTCHA v3

### 2. 配置后端

编辑 `middleware/recaptcha.go`：

```go
secretKey := "YOUR_RECAPTCHA_SECRET_KEY" // 替换为你的密钥
```

### 3. 配置前端

在数据库管理页面的 `getRecaptchaToken` 函数中集成 reCAPTCHA v3：

```javascript
const getRecaptchaToken = async (): Promise<string> => {
  return new Promise((resolve) => {
    grecaptcha.ready(async () => {
      const token = await grecaptcha.execute("YOUR_SITE_KEY", {
        action: "submit",
      });
      resolve(token);
    });
  });
};
```

### 开发模式

当前默认跳过验证（`secretKey === "YOUR_RECAPTCHA_SECRET_KEY"`时），生产环境请务必配置真实密钥。

## 文件清单

### 后端

- `initialize/db.go` - 数据库初始化（添加 is_admin 字段维护逻辑）
- `model/user.go` - 用户模型（添加 IsAdmin 字段）
- `api/user.go` - 用户接口（登录返回角色信息）
- `api/db_admin.go` - **新增**数据库管理 API
- `middleware/admin.go` - **新增**管理员权限中间件
- `middleware/recaptcha.go` - **新增**reCAPTCHA 验证中间件
- `router/router.go` - 路由注册（添加管理员接口路由）

### 前端

- `src/api/dbAdmin.ts` - **新增**数据库管理 API 封装
- `src/views/DbAdmin/index.vue` - **新增**数据库管理页面
- `src/router/index.ts` - 路由配置（添加权限控制）
- `src/views/Home/components/HeaderSection.vue` - 添加管理入口按钮

## 注意事项

1. **权限控制**：确保只授予可信用户管理员权限
2. **数据备份**：操作前请备份数据库
3. **验证码配置**：生产环境必须配置真实的 reCAPTCHA 密钥
4. **操作日志**：所有操作已记录到日志，便于审计
5. **SQL 注入防护**：已使用参数化查询，但仍需谨慎操作

## 测试建议

1. 创建测试用户，设置为管理员
2. 登录后访问数据库管理页面
3. 测试各种操作（查询、新增、编辑、删除）
4. 验证非管理员无法访问
5. 测试批量操作功能

## 后续优化建议

1. 集成真实的 Google reCAPTCHA v3
2. 添加操作审计日志表
3. 支持 SQL 执行历史记录
4. 添加数据导出功能（CSV/Excel）
5. 支持数据库备份/恢复
6. 添加字段类型编辑器（日期选择器、下拉框等）

---

**开发完成时间**：2025-12-06  
**版本**：v1.0
