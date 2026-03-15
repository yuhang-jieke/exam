package main

import (
	"context"
	"testing"
)

func Test_handleGoodsMessage(t *testing.T) {
	type args struct {
		ctx  context.Context
		body []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handleGoodsMessage(tt.args.ctx, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("handleGoodsMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_waitForShutdown(t *testing.T) {
	type args struct {
		cancel context.CancelFunc
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			waitForShutdown(tt.args.cancel)
		})
	}
}
