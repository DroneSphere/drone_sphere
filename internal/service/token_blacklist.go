package service

import (
	"errors"
	"sync"
	"time"
)

// TokenBlacklistService token黑名单服务接口
type TokenBlacklistService interface {
	// AddToken 将token加入黑名单
	AddToken(tokenString string) error
	// IsTokenBlacklisted 检查token是否在黑名单中
	IsTokenBlacklisted(tokenString string) bool
	// RemoveExpiredTokens 清理过期的token
	RemoveExpiredTokens()
}

// TokenBlacklistImpl token黑名单服务实现
type TokenBlacklistImpl struct {
	blacklist map[string]time.Time // token黑名单，存储token和过期时间
	mutex     sync.RWMutex         // 保护黑名单的并发访问
}

// NewTokenBlacklistService 创建token黑名单服务实例
func NewTokenBlacklistService() TokenBlacklistService {
	return &TokenBlacklistImpl{
		blacklist: make(map[string]time.Time),
	}
}

// AddToken 将token加入黑名单
func (s *TokenBlacklistImpl) AddToken(tokenString string) error {
	if tokenString == "" {
		return errors.New(ErrInvalidToken)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 将token加入黑名单，设置过期时间为当前时间+24小时（比JWT有效期稍长）
	expireTime := time.Now().Add(24 * time.Hour)
	s.blacklist[tokenString] = expireTime

	return nil
}

// IsTokenBlacklisted 检查token是否在黑名单中
func (s *TokenBlacklistImpl) IsTokenBlacklisted(tokenString string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	expireTime, exists := s.blacklist[tokenString]
	if !exists {
		return false
	}

	// 如果token已过期，从黑名单中移除
	if time.Now().After(expireTime) {
		delete(s.blacklist, tokenString)
		return false
	}

	return true
}

// RemoveExpiredTokens 清理过期的token
func (s *TokenBlacklistImpl) RemoveExpiredTokens() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	for token, expireTime := range s.blacklist {
		if now.After(expireTime) {
			delete(s.blacklist, token)
		}
	}
}