# API参考

<cite>
**本文档引用的文件**   
- [router.go](file://router/router.go)
- [user.go](file://api/user.go)
- [subject.go](file://api/subject.go)
- [category.go](file://api/category.go)
- [point.go](file://api/point.go)
- [question.go](file://api/question.go)
- [share.go](file://api/share.go)
- [common.go](file://api/common.go)
- [subject.ts](file://practice_problems_web/src/api/subject.ts)
- [category.ts](file://practice_problems_web/src/api/category.ts)
- [point.ts](file://practice_problems_web/src/api/point.ts)
- [question.ts](file://practice_problems_web/src/api/question.ts)
- [share.ts](file://practice_problems_web/src/api/share.ts)
- [index.ts](file://practice_problems_web/src/types/index.ts)
</cite>

## 目录
1. [简介](#简介)
2. [认证机制](#认证机制)
3. [用户管理](#用户管理)
4. [科目管理](#科目管理)
5. [分类管理](#分类管理)
6. [知识点管理](#知识点管理)
7. [题目管理](#题目管理)
8. [分享与授权](#分享与授权)
9. [文件上传](#文件上传)

## 简介
本API参考文档详细描述了基于Gin框架构建的RESTful API接口。所有API端点均以`/api/v1`为前缀，通过JWT进行身份验证。文档涵盖了用户认证、资源管理、权限控制等核心功能，为前端开发提供完整的接口规范。

**Section sources**
- [router.go](file://router/router.go#L33-L106)

## 认证机制
系统采用JWT（JSON Web Token）进行用户身份验证。所有需要认证的接口都必须在请求头中包含`Authorization: Bearer <token>`。

- **注册**：`POST /api/v1/auth/register` 创建新用户
- **登录**：`POST /api/v1/auth/login` 获取访问令牌
- **退出**：`POST /api/v1/auth/logout` 注销当前会话

成功登录后，服务器返回的响应中包含`token`字段，该令牌需在后续请求的`Authorization`头中使用。

**Section sources**
- [router.go](file://router/router.go#L39-L40)
- [user.go](file://api/user.go#L55-L259)

## 用户管理

### 用户注册
创建新用户账户。

**HTTP方法**：`POST`  
**URL路径**：`/api/v1/auth/register`  
**认证**：无需认证

**请求头**：
```
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "username": "string",
  "password": "string",
  "nickname": "string",
  "email": "string"
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "注册成功",
  "data": null
}
```

**HTTP状态码**：
- `200`：注册成功
- `400`：参数错误
- `500`：系统错误

**curl命令示例**：
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass",
    "nickname": "测试用户",
    "email": "test@example.com"
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/auth/register', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    username: 'testuser',
    password: 'testpass',
    nickname: '测试用户',
    email: 'test@example.com'
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L39)
- [user.go](file://api/user.go#L55-L95)

### 用户登录
获取访问令牌。

**HTTP方法**：`POST`  
**URL路径**：`/api/v1/auth/login`  
**认证**：无需认证

**请求头**：
```
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "username": "string",
  "password": "string"
}
```

**响应体JSON Schema**：
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

**HTTP状态码**：
- `200`：登录成功
- `402`：密码错误
- `404`：用户不存在
- `500`：系统错误

**curl命令示例**：
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass"
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    username: 'testuser',
    password: 'testpass'
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L40)
- [user.go](file://api/user.go#L100-L240)

### 修改用户信息
更新用户资料或修改密码。

**HTTP方法**：`PUT`  
**URL路径**：`/api/v1/user/profile`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "nickname": "string",
  "email": "string",
  "old_password": "string",
  "new_password": "string"
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "更新成功",
  "data": null
}
```

**HTTP状态码**：
- `200`：更新成功
- `400`：参数错误或旧密码错误
- `401`：未授权
- `500`：系统错误

**curl命令示例**：
```bash
curl -X PUT http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nickname": "新昵称",
    "email": "new@example.com",
    "old_password": "oldpass",
    "new_password": "newpass"
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/user/profile', {
  method: 'PUT',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    nickname: '新昵称',
    email: 'new@example.com',
    old_password: 'oldpass',
    new_password: 'newpass'
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L49)
- [user.go](file://api/user.go#L264-L341)

## 科目管理

### 获取科目列表
获取当前用户有权访问的科目列表。

**HTTP方法**：`GET`  
**URL路径**：`/api/v1/subjects`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "success",
  "data": [
    {
      "id": 1,
      "name": "数学",
      "status": 1,
      "creatorCode": "string",
      "createTime": "string",
      "updateTime": "string",
      "creatorEmail": "string",
      "creatorName": "string"
    }
  ]
}
```

**HTTP状态码**：
- `200`：获取成功
- `401`：未授权
- `500`：系统错误

**curl命令示例**：
```bash
curl -X GET http://localhost:8080/api/v1/subjects \
  -H "Authorization: Bearer <your_token>"
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/subjects', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ' + token
  }
})
```

**Section sources**
- [router.go](file://router/router.go#L71)
- [subject.go](file://api/subject.go#L19-L79)

### 创建科目
创建新的科目。

**HTTP方法**：`POST`  
**URL路径**：`/api/v1/subjects`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "name": "string",
  "status": 1
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "创建成功",
  "data": {
    "id": 1
  }
}
```

**HTTP状态码**：
- `200`：创建成功
- `400`：参数错误
- `401`：未授权
- `500`：系统错误

**curl命令示例**：
```bash
curl -X POST http://localhost:8080/api/v1/subjects \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "物理",
    "status": 1
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/subjects', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    name: '物理',
    status: 1
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L73)
- [subject.go](file://api/subject.go#L137-L185)

### 更新科目
修改现有科目的信息。

**HTTP方法**：`PUT`  
**URL路径**：`/api/v1/subjects/:id`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "name": "string",
  "status": 1
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "更新成功",
  "data": null
}
```

**HTTP状态码**：
- `200`：更新成功
- `400`：ID参数错误
- `403`：无权操作
- `404`：科目不存在
- `500`：系统错误

**curl命令示例**：
```bash
curl -X PUT http://localhost:8080/api/v1/subjects/1 \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "高级物理",
    "status": 1
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/subjects/1', {
  method: 'PUT',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    name: '高级物理',
    status: 1
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L74)
- [subject.go](file://api/subject.go#L187-L251)

### 删除科目
删除指定的科目。

**HTTP方法**：`DELETE`  
**URL路径**：`/api/v1/subjects/:id`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "删除成功",
  "data": null
}
```

**HTTP状态码**：
- `200`：删除成功
- `400`：ID参数错误
- `403`：无权操作
- `404`：科目不存在
- `500`：系统错误

**curl命令示例**：
```bash
curl -X DELETE http://localhost:8080/api/v1/subjects/1 \
  -H "Authorization: Bearer <your_token>"
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/subjects/1', {
  method: 'DELETE',
  headers: {
    'Authorization': 'Bearer ' + token
  }
})
```

**Section sources**
- [router.go](file://router/router.go#L75)
- [subject.go](file://api/subject.go#L253-L310)

## 分类管理

### 获取分类列表
获取指定科目的分类列表。

**HTTP方法**：`GET`  
**URL路径**：`/api/v1/categories`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
```

**查询参数**：
- `subject_id` (必需)：科目ID

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "success",
  "data": [
    {
      "id": 1,
      "subject_id": 1,
      "categorie_name": "代数",
      "create_time": "string",
      "update_time": "string",
      "sort_order": 0,
      "difficulty": 0
    }
  ]
}
```

**HTTP状态码**：
- `200`：获取成功
- `400`：缺少subject_id参数
- `401`：未授权
- `403`：无权访问
- `500`：系统错误

**curl命令示例**：
```bash
curl -X GET "http://localhost:8080/api/v1/categories?subject_id=1" \
  -H "Authorization: Bearer <your_token>"
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/categories?subject_id=1', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ' + token
  }
})
```

**Section sources**
- [router.go](file://router/router.go#L83)
- [category.go](file://api/category.go#L19-L86)

### 创建分类
在指定科目下创建新的分类。

**HTTP方法**：`POST`  
**URL路径**：`/api/v1/categories`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "subject_id": 1,
  "category_name": "几何"
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "创建成功",
  "data": {
    "id": 1
  }
}
```

**HTTP状态码**：
- `200`：创建成功
- `400`：参数错误
- `403`：无权操作
- `404`：科目不存在
- `500`：系统错误

**curl命令示例**：
```bash
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "subject_id": 1,
    "category_name": "几何"
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/categories', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    subject_id: 1,
    category_name: '几何'
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L84)
- [category.go](file://api/category.go#L91-L147)

### 更新分类
修改分类信息。

**HTTP方法**：`PUT`  
**URL路径**：`/api/v1/categories/:id`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "category_name": "string",
  "difficulty": 0
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "更新成功",
  "data": null
}
```

**HTTP状态码**：
- `200`：更新成功
- `400`：参数错误
- `403`：无权操作
- `404`：分类不存在
- `500`：系统错误

**curl命令示例**：
```bash
curl -X PUT http://localhost:8080/api/v1/categories/1 \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "category_name": "平面几何",
    "difficulty": 1
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/categories/1', {
  method: 'PUT',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    category_name: '平面几何',
    difficulty: 1
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L85)
- [category.go](file://api/category.go#L152-L224)

### 删除分类
删除指定的分类。

**HTTP方法**：`DELETE`  
**URL路径**：`/api/v1/categories/:id`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "删除成功",
  "data": null
}
```

**HTTP状态码**：
- `200`：删除成功
- `403`：无权操作
- `404`：分类不存在
- `500`：系统错误

**curl命令示例**：
```bash
curl -X DELETE http://localhost:8080/api/v1/categories/1 \
  -H "Authorization: Bearer <your_token>"
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/categories/1', {
  method: 'DELETE',
  headers: {
    'Authorization': 'Bearer ' + token
  }
})
```

**Section sources**
- [router.go](file://router/router.go#L86)
- [category.go](file://api/category.go#L229-L283)

## 知识点管理

### 获取知识点列表
获取指定分类下的知识点列表。

**HTTP方法**：`GET`  
**URL路径**：`/api/v1/points`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
```

**查询参数**：
- `category_id` (必需)：分类ID

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "success",
  "data": [
    {
      "id": 1,
      "title": "二次方程",
      "create_time": "string",
      "sort_order": 0,
      "difficulty": 0
    }
  ]
}
```

**HTTP状态码**：
- `200`：获取成功
- `400`：缺少category_id参数
- `401`：未授权
- `403`：无权访问
- `500`：系统错误

**curl命令示例**：
```bash
curl -X GET "http://localhost:8080/api/v1/points?category_id=1" \
  -H "Authorization: Bearer <your_token>"
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/points?category_id=1', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ' + token
  }
})
```

**Section sources**
- [router.go](file://router/router.go#L90)
- [point.go](file://api/point.go#L19-L79)

### 获取知识点详情
获取指定知识点的详细信息。

**HTTP方法**：`GET`  
**URL路径**：`/api/v1/points/:id`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "id": 1,
    "category_id": 1,
    "title": "二次方程",
    "content": "详细内容",
    "reference_links": "[]",
    "local_image_names": "[]",
    "create_time": "string",
    "update_time": "string"
  }
}
```

**HTTP状态码**：
- `200`：获取成功
- `403`：无权查看
- `404`：知识点不存在
- `500`：系统错误

**curl命令示例**：
```bash
curl -X GET http://localhost:8080/api/v1/points/1 \
  -H "Authorization: Bearer <your_token>"
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/points/1', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ' + token
  }
})
```

**Section sources**
- [router.go](file://router/router.go#L91)
- [point.go](file://api/point.go#L84-L132)

### 创建知识点
在指定分类下创建新的知识点。

**HTTP方法**：`POST`  
**URL路径**：`/api/v1/points`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "category_id": 1,
  "title": "新知识点"
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "创建成功",
  "data": {
    "id": 1
  }
}
```

**HTTP状态码**：
- `200`：创建成功
- `400`：参数错误
- `403`：无权操作
- `404`：分类不存在
- `500`：系统错误

**curl命令示例**：
```bash
curl -X POST http://localhost:8080/api/v1/points \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": 1,
    "title": "三角函数"
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/points', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    category_id: 1,
    title: '三角函数'
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L92)
- [point.go](file://api/point.go#L135-L193)

### 更新知识点
修改知识点信息。

**HTTP方法**：`PUT`  
**URL路径**：`/api/v1/points/:id`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "title": "string",
  "content": "string",
  "reference_links": "string",
  "local_image_names": "string",
  "difficulty": 0,
  "category_id": 1
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "更新成功",
  "data": null
}
```

**HTTP状态码**：
- `200`：更新成功
- `400`：参数错误
- `403`：无权操作
- `404`：知识点不存在
- `500`：系统错误

**curl命令示例**：
```bash
curl -X PUT http://localhost:8080/api/v1/points/1 \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "高级三角函数",
    "content": "详细内容...",
    "difficulty": 2
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/points/1', {
  method: 'PUT',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    title: '高级三角函数',
    content: '详细内容...',
    difficulty: 2
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L93)
- [point.go](file://api/point.go#L198-L313)

## 题目管理

### 获取题目列表
获取指定知识点或分类下的题目列表。

**HTTP方法**：`GET`  
**URL路径**：`/api/v1/questions`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
```

**查询参数**（二选一）：
- `point_id`：知识点ID
- `category_id`：分类ID

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "success",
  "data": [
    {
      "id": 1,
      "knowledge_point_id": 1,
      "question_text": "题目内容",
      "option1": "选项A",
      "option2": "选项B",
      "option3": "选项C",
      "option4": "选项D",
      "correct_answer": 1,
      "explanation": "解析",
      "note": "用户备注",
      "create_time": "string"
    }
  ]
}
```

**HTTP状态码**：
- `200`：获取成功
- `400`：缺少必要参数
- `401`：未授权
- `403`：无权访问
- `500`：系统错误

**curl命令示例**：
```bash
curl -X GET "http://localhost:8080/api/v1/questions?point_id=1" \
  -H "Authorization: Bearer <your_token>"
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/questions?point_id=1', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ' + token
  }
})
```

**Section sources**
- [router.go](file://router/router.go#L99)
- [question.go](file://api/question.go#L17-L177)

### 创建题目
在指定知识点下创建新题目。

**HTTP方法**：`POST`  
**URL路径**：`/api/v1/questions`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "knowledge_point_id": 1,
  "question_text": "题目内容",
  "option1": "选项A",
  "option2": "选项B",
  "option3": "选项C",
  "option4": "选项D",
  "correct_answer": 1,
  "explanation": "解析",
  "note": "初始备注"
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "创建成功",
  "data": {
    "id": 1
  }
}
```

**HTTP状态码**：
- `200`：创建成功
- `400`：参数错误
- `403`：无权操作
- `404`：知识点不存在
- `500`：系统错误

**curl命令示例**：
```bash
curl -X POST http://localhost:8080/api/v1/questions \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "knowledge_point_id": 1,
    "question_text": "2+2=?",
    "option1": "3",
    "option2": "4",
    "option3": "5",
    "option4": "6",
    "correct_answer": 2,
    "explanation": "基本算术"
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/questions', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    knowledge_point_id: 1,
    question_text: '2+2=?',
    option1: '3',
    option2: '4',
    option3: '5',
    option4: '6',
    correct_answer: 2,
    explanation: '基本算术'
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L100)
- [question.go](file://api/question.go#L183-L248)

### 更新题目
修改现有题目的信息。

**HTTP方法**：`PUT`  
**URL路径**：`/api/v1/questions/:id`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "question_text": "string",
  "option1": "string",
  "option2": "string",
  "option3": "string",
  "option4": "string",
  "correct_answer": 1,
  "explanation": "string"
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "更新成功",
  "data": null
}
```

**HTTP状态码**：
- `200`：更新成功
- `400`：参数错误
- `403`：无权操作
- `404`：题目不存在
- `500`：系统错误

**curl命令示例**：
```bash
curl -X PUT http://localhost:8080/api/v1/questions/1 \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "question_text": "2+2=?",
    "option1": "3",
    "option2": "4",
    "option3": "5",
    "option4": "6",
    "correct_answer": 2,
    "explanation": "基本算术运算"
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/questions/1', {
  method: 'PUT',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    question_text: '2+2=?',
    option1: '3',
    option2: '4',
    option3: '5',
    option4: '6',
    correct_answer: 2,
    explanation: '基本算术运算'
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L101)
- [question.go](file://api/question.go#L253-L325)

### 修改用户题目备注
更新用户对特定题目的个人备注。

**HTTP方法**：`POST`  
**URL路径**：`/api/v1/questions/note`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "question_id": 1,
  "note": "我的学习笔记..."
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "备注已保存",
  "data": null
}
```

**HTTP状态码**：
- `200`：保存成功
- `400`：参数错误
- `401`：未授权
- `500`：系统错误

**curl命令示例**：
```bash
curl -X POST http://localhost:8080/api/v1/questions/note \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "question_id": 1,
    "note": "这是一道重要题目，需要重点复习"
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/questions/note', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    question_id: 1,
    note: '这是一道重要题目，需要重点复习'
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L103)
- [question.go](file://api/question.go#L330-L376)

## 分享与授权

### 创建分享
创建分享码或直接授权给指定用户。

**HTTP方法**：`POST`  
**URL路径**：`/api/v1/share/create`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "subject_ids": [1, 2],
  "duration": "3d",
  "type": 0,
  "targets": []
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "生成成功",
  "data": "SHARE-ABC123"
}
```

**HTTP状态码**：
- `200`：创建成功
- `400`：参数错误
- `403`：科目非本人所有
- `500`：系统错误

**curl命令示例**：
```bash
curl -X POST http://localhost:8080/api/v1/share/create \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "subject_ids": [1],
    "duration": "7d",
    "type": 0
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/share/create', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    subject_ids: [1],
    duration: '7d',
    type: 0
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L64)
- [share.go](file://api/share.go#L53-L132)

### 绑定分享
使用分享码绑定资源。

**HTTP方法**：`POST`  
**URL路径**：`/api/v1/share/bind`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体JSON Schema**：
```json
{
  "code": "SHARE-ABC123"
}
```

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "成功绑定 1 个新科目！",
  "data": {
    "success_count": 1,
    "skipped_count": 0,
    "total_users": 5
  }
}
```

**HTTP状态码**：
- `200`：绑定成功
- `400`：参数错误或分享码已失效
- `404`：分享码无效
- `500`：系统错误

**curl命令示例**：
```bash
curl -X POST http://localhost:8080/api/v1/share/bind \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "SHARE-ABC123"
  }'
```

**JavaScript/TypeScript调用示例**：
```typescript
fetch('/api/v1/share/bind', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    code: 'SHARE-ABC123'
  })
})
```

**Section sources**
- [router.go](file://router/router.go#L65)
- [share.go](file://api/share.go#L247-L406)

## 文件上传

### 上传图片
上传图片文件到服务器。

**HTTP方法**：`POST`  
**URL路径**：`/api/v1/upload`  
**认证**：需要JWT认证

**请求头**：
```
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**表单数据**：
- `file`：文件对象
- `type`：业务类型（point/question）
- `pointId`：知识点ID（当type为point时）

**响应体JSON Schema**：
```json
{
  "code": 200,
  "msg": "上传成功",
  "data": {
    "url": "/uploads/point/20231201/uuid.jpg",
    "path": "/uploads/point/20231201/uuid.jpg"
  }
}
```

**HTTP状态码**：
- `200`：上传成功
- `400`：参数错误或文件过大
- `401`：未授权
- `403`：无权操作
- `500`：系统错误

**curl命令示例**：
```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer <your_token>" \
  -F "file=@/path/to/image.jpg" \
  -F "type=point" \
  -F "pointId=1"
```

**JavaScript/TypeScript调用示例**：
```typescript
const formData = new FormData();
formData.append('file', file);
formData.append('type', 'point');
formData.append('pointId', '1');

fetch('/api/v1/upload', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token
  },
  body: formData
})
```

**Section sources**
- [router.go](file://router/router.go#L53)
- [common.go](file://api/common.go#L26-L174)