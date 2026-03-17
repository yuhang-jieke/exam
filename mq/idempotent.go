package mq

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// IdempotentHandler 幂等性处理器（基于 Redis SETNX）
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

// TryAcquire 尝试获取消息处理权（使用 Redis SETNX）
// 返回 true 表示获取成功，可以处理消息
// 返回 false 表示消息已被其他消费者处理，应跳过
func (h *IdempotentHandler) TryAcquire(ctx context.Context, messageKey string) (bool, error) {
	key := h.keyPrefix + messageKey

	// 使用 SetNX (Set if Not eXists) 原子操作
	// 返回 true = 设置成功（key不存在），可以处理消息
	// 返回 false = key已存在，消息已被处理，跳过
	ok, err := h.rdb.SetNX(ctx, key, "1", h.expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis SETNX failed: %w", err)
	}

	return ok, nil
}

// IsProcessed 检查消息是否已处理
// 返回 true=已处理过（跳过），false=未处理过（继续处理）
// 注意：此方法会同时设置标记，如需只检查不设置，使用 CheckExists
func (h *IdempotentHandler) IsProcessed(ctx context.Context, messageKey string) (bool, error) {
	acquired, err := h.TryAcquire(ctx, messageKey)
	if err != nil {
		return false, err
	}
	// acquired=true 表示未处理过，所以返回 false
	// acquired=false 表示已处理过，所以返回 true
	return !acquired, nil
}

// CheckExists 仅检查是否存在（不设置标记）
func (h *IdempotentHandler) CheckExists(ctx context.Context, messageKey string) (bool, error) {
	key := h.keyPrefix + messageKey
	count, err := h.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Release 释放处理权（处理失败时调用，允许重试）
func (h *IdempotentHandler) Release(ctx context.Context, messageKey string) error {
	key := h.keyPrefix + messageKey
	return h.rdb.Del(ctx, key).Err()
}

// MarkProcessed 标记消息为已处理
func (h *IdempotentHandler) MarkProcessed(ctx context.Context, messageKey string) error {
	key := h.keyPrefix + messageKey
	return h.rdb.Set(ctx, key, "1", h.expiration).Err()
}

// RemoveProcessed 移除已处理标记（用于重试场景）
func (h *IdempotentHandler) RemoveProcessed(ctx context.Context, messageKey string) error {
	return h.Release(ctx, messageKey)
}

// GenerateMessageKey 生成消息唯一标识（基于消息内容 MD5 Hash）
func GenerateMessageKey(body []byte) string {
	hash := md5.Sum(body)
	return hex.EncodeToString(hash[:])
}

// GenerateMessageKeyWithFields 基于指定字段生成消息唯一标识
func GenerateMessageKeyWithFields(fields ...string) string {
	data := ""
	for _, f := range fields {
		data += f + ":"
	}
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ExecuteWithIdempotent 带幂等性检查执行处理函数（推荐使用）
// 自动处理成功/失败场景，处理失败会释放标记允许重试
func (h *IdempotentHandler) ExecuteWithIdempotent(ctx context.Context, messageKey string, handler func() error) error {
	// 1. 尝试获取处理权
	acquired, err := h.TryAcquire(ctx, messageKey)
	if err != nil {
		return fmt.Errorf("idempotent check failed: %w", err)
	}

	if !acquired {
		// 已被其他消费者处理，跳过
		return nil
	}

	// 2. 执行业务处理
	if err := handler(); err != nil {
		// 处理失败，释放标记以便重试
		if releaseErr := h.Release(ctx, messageKey); releaseErr != nil {
			return fmt.Errorf("handler failed: %v, release failed: %w", err, releaseErr)
		}
		return err
	}

	// 3. 处理成功，标记会自动过期
	return nil
}

// ExecuteWithIdempotentBody 基于消息体的幂等性处理（自动生成 key）
func (h *IdempotentHandler) ExecuteWithIdempotentBody(ctx context.Context, body []byte, handler func() error) error {
	messageKey := GenerateMessageKey(body)
	return h.ExecuteWithIdempotent(ctx, messageKey, handler)
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

// OrderIdempotentHandler 订单专用幂等性处理器（7天过期）
func OrderIdempotentHandler(rdb *redis.Client) *IdempotentHandler {
	return NewIdempotentHandler(rdb, "order:processed:", 7*24*time.Hour)
}

// GoodsIdempotentHandler 商品专用幂等性处理器（24小时过期）
func GoodsIdempotentHandler(rdb *redis.Client) *IdempotentHandler {
	return NewIdempotentHandler(rdb, "goods:processed:", 24*time.Hour)
}
