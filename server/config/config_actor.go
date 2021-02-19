// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"context"
	"fmt"

	"github.com/pole-group/pole/common"
)

type HandlerChain struct {
	filters []ConfigFilter
}

func (fc *HandlerChain) Init(ctx *common.ContextPole) {
	fc.filters = append(fc.filters, &CapacityConfigFilter{})
	fc.filters = append(fc.filters, &AuditConfigFilter{})
	fc.filters = append(fc.filters, &EncryptConfigFilter{})
}

//Do 对请求做处理
func (fc *HandlerChain) Do(ctx context.Context, cfg interface{}) error {
	switch c := cfg.(type) {
	case *ConfigFile:
		for _, filter := range fc.filters {
			filter.doFilter(ctx, c)
		}
	case *ConfigBetaFile:
		for _, filter := range fc.filters {
			filter.doBetaFilter(ctx, c)
		}
	default:
		return fmt.Errorf("unsupport config implement")
	}
	return nil
}

//Shutdown 关闭操作链
func (fc *HandlerChain) Shutdown() {

}

type ConfigFilter interface {
	// doFilter
	doFilter(ctx context.Context, cfg *ConfigFile)
	// doBetaFilter
	doBetaFilter(ctx context.Context, cfg *ConfigBetaFile)
	// doHistoryFilter
	doHistoryFilter(ctx context.Context, cfg *ConfigHistoryFile)
	// doOnFinish
	doOnFinish(ctx context.Context, cfg ConfigMetadata)
}

// 容量管理记录
type CapacityConfigFilter struct {
}

func (cf *CapacityConfigFilter) doFilter(ctx context.Context, cfg *ConfigFile) {

}

func (cf *CapacityConfigFilter) doBetaFilter(ctx context.Context, cfg *ConfigBetaFile) {

}

func (cf *CapacityConfigFilter) doHistoryFilter(ctx context.Context, cfg *ConfigHistoryFile) {

}

func (cf *CapacityConfigFilter) doOnFinish(ctx context.Context, cfg ConfigMetadata) {

}

// 配置操作审计记录
type AuditConfigFilter struct {
}

func (af *AuditConfigFilter) doFilter(ctx context.Context, cfg *ConfigFile) {

}

func (af *AuditConfigFilter) doBetaFilter(ctx context.Context, cfg *ConfigBetaFile) {

}

func (af *AuditConfigFilter) doHistoryFilter(ctx context.Context, cfg *ConfigHistoryFile) {

}

func (af *AuditConfigFilter) doOnFinish(ctx context.Context, cfg ConfigMetadata) {

}

// 配置加密操作
type EncryptConfigFilter struct {
}

func (ef *EncryptConfigFilter) doFilter(ctx context.Context, cfg *ConfigFile) {

}

func (ef *EncryptConfigFilter) doBetaFilter(ctx context.Context, cfg *ConfigBetaFile) {

}

func (ef *EncryptConfigFilter) doHistoryFilter(ctx context.Context, cfg *ConfigHistoryFile) {

}

func (ef *EncryptConfigFilter) doOnFinish(ctx context.Context, cfg ConfigMetadata) {

}
