package config

import (
	"fmt"

	"github.com/pole-group/pole/common"
)

type FilterChain struct {
	filters []ConfigFilter
}

func (fc *FilterChain) Init(ctx *common.ContextPole) {
	fc.filters = append(fc.filters, &CapacityConfigFilter{})
	fc.filters = append(fc.filters, &AuditConfigFilter{})
	fc.filters = append(fc.filters, &EncryptConfigFilter{})
}

func (fc *FilterChain) Do(ctx common.ContextPole, cfg interface{}) error {
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

func (fc *FilterChain) Shutdown()  {

}


type ConfigFilter interface {
	doFilter(ctx common.ContextPole, cfg *ConfigFile)

	doBetaFilter(ctx common.ContextPole, cfg *ConfigBetaFile)

	doOnFinish(ctx common.ContextPole, cfg ConfigMetadata)
}

// 容量管理记录
type CapacityConfigFilter struct {
}

func (cf *CapacityConfigFilter) doFilter(ctx common.ContextPole, cfg *ConfigFile) {

}

func (cf *CapacityConfigFilter) doBetaFilter(ctx common.ContextPole, cfg *ConfigBetaFile) {

}

func (cf *CapacityConfigFilter) doOnFinish(ctx common.ContextPole, cfg ConfigMetadata) {

}

// 配置操作审计记录
type AuditConfigFilter struct {
}

func (af *AuditConfigFilter) doFilter(ctx common.ContextPole, cfg *ConfigFile) {

}

func (af *AuditConfigFilter) doBetaFilter(ctx common.ContextPole, cfg *ConfigBetaFile) {

}

func (af *AuditConfigFilter) doOnFinish(ctx common.ContextPole, cfg ConfigMetadata) {

}

type EncryptConfigFilter struct {
}

func (ef *EncryptConfigFilter) doFilter(ctx common.ContextPole, cfg *ConfigFile) {

}

func (ef *EncryptConfigFilter) doBetaFilter(ctx common.ContextPole, cfg *ConfigBetaFile) {

}

func (ef *EncryptConfigFilter) doOnFinish(ctx common.ContextPole, cfg ConfigMetadata) {

}
