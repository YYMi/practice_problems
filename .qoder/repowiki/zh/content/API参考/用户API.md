# 用户API

<cite>
**本文档引用的文件**
- [user.go](file://api/user.go)
- [model/user.go](file://model/user.go)
- [router.go](file://router/router.go)
- [jwt.go](file://middleware/jwt.go)
- [token_store.go](file://global/token_store.go)
</cite>

## 目录
1. [用户API概述](#用户api概述)
2. [接口详细说明](#接口详细说明)
   - [POST /api/v1/auth/register（用户注册）](#post-apiv1authregister用户注册)
   - [POST /api/v1/auth/login（用户登录）](#post-apiv1authlogin用户登录)
   - [PUT /api/v1/user/profile（更新用户信息或密码）](#put-apiv1userprofile更新用户信息或密码)
   - [POST /api/v1/auth/logout（用户登出）](#post-apiv1authlogout用户登出)
3. [安全注意事项](#安全注意事项)
4. [调用示例](#调用示例)

## 用户API概述

本文档详细说明了用户相关API接口的使用方法，包括用户注册、登录、信息更新和登出等功能。系统采用JWT（JSON Web Token）进行身份认证，所有需要认证的接口都要求在请求头中包含有效的Authorization令牌。

**Section sources**
- [router.go](file://router/router.go#L38-L50)

## 接口详细说明

### POST /api/v1/auth/register（用户注册）

**接口说明**：创建新用户账户

**HTTP方法**：POST

**完整URL**：`/api/v1/auth/register`

**认证要求**：公开接口，无需JWT认证

**请求头**：
- `Content-Type: application/json`

**请求体**：
```json
{
  "username": "string",
  "password": "string",
  "nickname": "string",
  "email": "string"
}
```

**字段说明**：
- `username`：用户名，必填，字符串类型
- `password`：密码，必填，字符串类型
- `nickname`：昵称，可选，字符串类型
- `email`：邮箱，可选，字符串类型

**校验规则**：
- username和password为必填字段
- 用户名不能重复

**成功响应**：
```json
{
  "code": 200,
  "msg": "注册成功",
  "data": null
}
```

**失败响应**：
```json
{
  "code": 400,
  "msg": "参数错误: ..."
}
```
或
```json
{
  "code": 500,
  "msg": "注册失败，用户名可能已存在"
}
```

**Section sources**
- [user.go](file://api/user.go#L55-L95)
- [model/user.go](file://model/user.go#L6-L11)

### POST /api/v1/auth/login（用户登录）

**接口说明**：用户登录系统，支持两种登录方式：Token自动登录和账号密码登录

**HTTP方法**：POST

**完整URL**：`/api/v1/auth/login`

**认证要求**：公开接口，无需JWT认证

**请求头**：
- `Content-Type: application/json`

**请求体**：
```json
{
  "username": "string",
  "password": "string"
}
```

**字段说明**：
- `username`：用户名，必填，字符串类型
- `password`：密码，可为空，字符串类型

**特殊逻辑**：
- 系统首先检查请求头中的Authorization令牌，如果有效则自动登录
- 如果没有令牌或令牌无效，则尝试使用账号密码登录
- 支持空密码登录：如果用户账户的密码为空，则允许登录但返回`need_change_pwd: true`，提示前端需要强制修改密码

**成功响应**：
```json
{
  "code": 200,
  "msg": "登录成功",
  "data": {
    "token": "string",
    "user_code": "string",
    "username": "string",
    "nickname": "string",
    "email": "string",
    "need_change_pwd": false
  }
}
```

**失败响应**：
```json
{
  "code": 404,
  "msg": "用户不存在"
}
```
或
```json
{
  "code": 402,
  "msg": "密码错误"
}
```

**Section sources**
- [user.go](file://api/user.go#L100-L240)
- [model/user.go](file://model/user.go#L14-L17)

### PUT /api/v1/user/profile（更新用户信息或密码）

**接口说明**：更新用户个人信息或修改密码

**HTTP方法**：PUT

**完整URL**：`/api/v1/user/profile`

**认证要求**：需要JWT认证

**请求头**：
- `Content-Type: application/json`
- `Authorization: Bearer <token>`

**请求体**：
```json
{
  "nickname": "string",
  "email": "string",
  "old_password": "string",
  "new_password": "string"
}
```

**字段说明**：
- `nickname`：新昵称，可选
- `email`：新邮箱，可选
- `old_password`：旧密码，修改密码时必填
- `new_password`：新密码，修改密码时必填

**业务逻辑**：
- 如果`new_password`不为空，则执行密码修改逻辑
- 密码修改分两种情况：
  1. 用户原密码为空：直接允许设置新密码（初始设置密码流程）
  2. 用户原密码存在：必须提供正确的`old_password`进行验证
- 如果`nickname`或`email`不为空，则更新相应的用户信息

**成功响应**：
```json
{
  "code": 200,
  "msg": "更新成功",
  "data": null
}
```

**失败响应**：
```json
{
  "code": 401,
  "msg": "未授权"
}
```
或
```json
{
  "code": 400,
  "msg": "旧密码错误"
}
```

**Section sources**
- [user.go](file://api/user.go#L264-L341)
- [model/user.go](file://model/user.go#L20-L25)

### POST /api/v1/auth/logout（用户登出）

**接口说明**：用户退出登录，使当前Token失效

**HTTP方法**：POST

**完整URL**：`/api/v1/auth/logout`

**认证要求**：需要JWT认证

**请求头**：
- `Authorization: Bearer <token>`

**请求体**：无

**业务逻辑**：
- 从请求头获取Token
- 将Token从内存白名单中移除，使其失效
- 如果请求头中没有Token，也返回成功

**成功响应**：
```json
{
  "code": 200,
  "msg": "退出成功",
  "data": null
}
```

**Section sources**
- [user.go](file://api/user.go#L245-L259)
- [token_store.go](file://global/token_store.go#L32-L37)

## 安全注意事项

1. **密码安全**：
   - 系统采用双重保护机制：前端MD5 + 后端MD5 + Bcrypt加密
   - 密码存储流程：前端MD5 → 后端MD5 → Bcrypt → 数据库
   - 即使数据库泄露，攻击者也难以破解用户密码

2. **Token管理**：
   - 使用内存白名单机制管理有效Token
   - 登出时立即将Token从白名单中移除，确保即时失效
   - Token有效期为30天

3. **空密码处理**：
   - 支持空密码登录，用于新用户初始登录场景
   - 空密码登录成功后返回`need_change_pwd: true`，提示前端强制修改密码
   - 修改密码时，如果原密码为空则无需验证旧密码

4. **认证中间件**：
   - 所有需要认证的接口都通过JWTAuthMiddleware进行保护
   - 中间件会验证Token的有效性并从白名单中核对

**Section sources**
- [jwt.go](file://middleware/jwt.go)
- [token_store.go](file://global/token_store.go)

## 调用示例

### curl命令示例

**用户注册**：
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "nickname": "测试用户",
    "email": "test@example.com"
  }'
```

**用户登录**：
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }'
```

**更新用户信息**：
```bash
curl -X PUT http://localhost:8080/api/v1/user/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "nickname": "新昵称",
    "email": "newemail@example.com"
  }'
```

**修改密码**：
```bash
curl -X PUT http://localhost:8080/api/v1/user/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "old_password": "testpass123",
    "new_password": "newpass456"
  }'
```

**用户登出**：
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 前端调用代码片段

```javascript
// 封装的请求函数（基于axios）
const service = axios.create({
  baseURL: '/api/v1',
  timeout: 5000
});

// 请求拦截器：自动添加Token
service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 注册用户
export const register = (userData) => {
  return service.post('/auth/register', userData);
};

// 用户登录
export const login = (credentials) => {
  return service.post('/auth/login', credentials);
};

// 更新用户信息
export const updateUserProfile = (profileData) => {
  return service.put('/user/profile', profileData);
};

// 用户登出
export const logout = () => {
  return service.post('/auth/logout');
};
```

**Section sources**
- [request.ts](file://practice_problems_web/src/utils/request.ts)