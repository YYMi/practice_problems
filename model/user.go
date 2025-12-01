package model

import "database/sql"

// 注册请求结构
type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

// 登录请求结构
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password"` // 允许为空传
}

// 修改用户请求结构
type UpdateUserReq struct {
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
	OldPassword string `json:"old_password"` // 修改密码时必填
	NewPassword string `json:"new_password"` // 修改密码时必填
}

type DbUser struct {
	Id       int
	Username string
	Password string
	UserCode string
	Nickname sql.NullString
	Email    sql.NullString
}
