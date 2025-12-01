package global

import (
	"sync"
)

// TokenStore 用于存储所有“有效”的 Token
// Key: Token字符串, Value: UserCode (这样通过Token能直接反查到人)
var TokenStore = struct {
	sync.RWMutex
	Data map[string]string
}{
	Data: make(map[string]string),
}

// SaveToken 保存 Token (登录时调用)
func SaveToken(token string, userCode string) {
	TokenStore.Lock()
	defer TokenStore.Unlock()
	TokenStore.Data[token] = userCode
}

// VerifyToken 校验 Token 是否存在 (中间件调用)
// 返回: (是否存在, userCode)
func VerifyToken(token string) (bool, string) {
	TokenStore.RLock()
	defer TokenStore.RUnlock()
	userCode, exists := TokenStore.Data[token]
	return exists, userCode
}

// RemoveToken 删除 Token (退出时调用)
func RemoveToken(token string) {
	TokenStore.Lock()
	defer TokenStore.Unlock()
	delete(TokenStore.Data, token)
}

// ClearUserTokens (可选) 踢掉某个用户的所有 Token
// 这是一个高级功能，如果你想实现“单点登录”，可以在 SaveToken 前调用这个
func ClearUserTokens(targetUserCode string) {
	TokenStore.Lock()
	defer TokenStore.Unlock()
	for t, code := range TokenStore.Data {
		if code == targetUserCode {
			delete(TokenStore.Data, t)
		}
	}
}
