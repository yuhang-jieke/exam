package main

import (
	"testing"
)

func Test_handleGoodsMessage(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		wantErr bool
	}{
		// TODO: Add test cases
		// 需要真实的 Redis 连接才能测试
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handleGoodsMessage(tt.body); (err != nil) != tt.wantErr {
				t.Errorf("handleGoodsMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_handleOrderMessage(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		wantErr bool
	}{
		// TODO: Add test cases
		// 需要真实的 Redis 连接才能测试
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handleOrderMessage(tt.body); (err != nil) != tt.wantErr {
				t.Errorf("handleOrderMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
