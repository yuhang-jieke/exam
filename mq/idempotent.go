package mq

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// IdempotentHandler 幂等性处理器
type IdempotentHandler struct {
	rdb        *redis.Client
	keyPrefix  string        // Redis key 前缀
	expiration time.Duration // 过期时间
}

// NewIdempotentHandler 创建幂等性处理器
func NewIdempotentHandler(rdb *redis.Client, keyPrefix string, expiration time.Duration) *IdempotentHandler {
	return &IdempotentHandler{
		rdb:        rdb,
		keyPrefix:  keyPrefix,
		expiration: expiration,
	}
}

// IsProcessed 检查消息是否已处理（幂等性检查）
// messageKey: 消息唯一标识（如消息ID、消息内容的Hash等）
// 返回: true=已处理过（跳过），false=未处理过（继续处理）
func (h *IdempotentHandler) IsProcessed(ctx context.Context, messageKey string) (bool, error) {
	key := h.keyPrefix + messageKey

	// 使用 SETNX (Set if Not eXists) 原子操作
	// 返回 true 表示设置成功（未处理过）
	// 返回 false 表示已存在（已处理过）
	result, err := h.rdb.SetNX(ctx, key, "1", h.expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx failed: %w", err)
	}

	// result = true 表示设置成功，即消息未处理过
	// result = false 表示 key 已存在，即消息已处理过
	return !result, nil
}

// MarkProcessed 标记消息为已处理
func (h *IdempotentHandler) MarkProcessed(ctx context.Context, messageKey string) error {
	key := h.keyPrefix + messageKey
	return h.rdb.Set(ctx, key, "1", h.expiration).Err()
}

// RemoveProcessed 移除已处理标记（用于重试场景）
func (h *IdempotentHandler) RemoveProcessed(ctx context.Context, messageKey string) error {
	key := h.keyPrefix + messageKey
	return h.rdb.Del(ctx, key).Err()
}

// GenerateMessageKey 生成消息唯一标识（基于消息内容Hash）
func GenerateMessageKey(body []byte) string {
	hash := md5.Sum(body)
	return hex.EncodeToString(hash[:])
}

// GenerateMessageKeyWithField 基于指定字段生成消息唯一标识
func GenerateMessageKeyWithField(fields ...string) string {
	data := ""
	for _, f := range fields {
		data += f + ":"
	}
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ================== 便捷函数 ==================

// CheckAndMark 检查并标记（原子操作）
// 返回: true=已处理过（跳过），false=未处理过（已标记，继续处理）
func (h *IdempotentHandler) CheckAndMark(ctx context.Context, messageKey string) (bool, error) {
	return h.IsProcessed(ctx, messageKey)
}

// WithIdempotent 包装处理函数，自动进行幂等性检查
// handler: 业务处理函数
// keyFunc: 从消息体生成唯一key的函数
func (h *IdempotentHandler) WithIdempotent(ctx context.Context, body []byte, handler func() error) error {
	// 生成消息key
	messageKey := GenerateMessageKey(body)

	// 检查是否已处理
	processed, err := h.IsProcessed(ctx, messageKey)
	if err != nil {
		return fmt.Errorf("idempotent check failed: %w", err)
	}

	if processed {
		return nil // 已处理过，跳过
	}

	// 执行业务处理
	if err := handler(); err != nil {
		// 处理失败，移除标记以便重试
		h.RemoveProcessed(ctx, messageKey)
		return err
	}

	return nil
}

// WithIdempotentKey 包装处理函数，使用自定义key进行幂等性检查
func (h *IdempotentHandler) WithIdempotentKey(ctx context.Context, messageKey string, handler func() error) error {
	// 检查是否已处理
	processed, err := h.IsProcessed(ctx, messageKey)
	if err != nil {
		return fmt.Errorf("idempotent check failed: %w", err)
	}

	if processed {
		return nil // 已处理过，跳过
	}

	// 执行业务处理
	if err := handler(); err != nil {
		// 处理失败，移除标记以便重试
		h.RemoveProcessed(ctx, messageKey)
		return err
	}

	return nil
}

// ================== 预设配置 ==================

// DefaultIdempotentHandler 默认幂等性处理器（24小时过期）
func DefaultIdempotentHandler(rdb *redis.Client) *IdempotentHandler {
	return NewIdempotentHandler(rdb, "mq:processed:", 24*time.Hour)
}

// ShortIdempotentHandler 短期幂等性处理器（1小时过期）
func ShortIdempotentHandler(rdb *redis.Client) *IdempotentHandler {
	return NewIdempotentHandler(rdb, "mq:processed:", 1*time.Hour)
}

// LongIdempotentHandler 长期幂等性处理器（7天过期）
func LongIdempotentHandler(rdb *redis.Client) *IdempotentHandler {
	return NewIdempotentHandler(rdb, "mq:processed:", 7*24*time.Hour)
}
