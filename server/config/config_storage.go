// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"context"
	"database/sql"
	"github.com/pole-group/pole/plugin"
	"github.com/pole-group/pole/server/storage"
)

type CfgFunction func(tx *sql.Tx, cfg *ConfigFile)

type CbCfgFunction func(tx *sql.Tx, cfg *ConfigBetaFile)

type ConfigQuery struct {
	Namespace string
	Group     string
	FileName  string
}

type ConfigOpContext struct {
	ctx context.Context
	db  *sql.DB
	tx  *sql.Tx
}

func NewConfigOpContext(ctx context.Context, db *sql.DB) *ConfigOpContext {
	return &ConfigOpContext{
		ctx: ctx,
		db:  db,
	}
}

func (ctx *ConfigOpContext) Begin() (bool, error) {
	tx, err := ctx.db.BeginTx(ctx.ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return false, err
	}
	ctx.tx = tx
	return true, nil
}

func (ctx *ConfigOpContext) Run(cfg *ConfigFile, chain ...CfgFunction) {
	for _, op := range chain {
		op(ctx.tx, cfg)
	}
}

func (ctx *ConfigOpContext) RunBeta(cfg *ConfigBetaFile, chain ...[]CfgFunction) {
	for _, op := range chain {
		op(ctx.tx, cfg)
	}
}

func (ctx *ConfigOpContext) Commit() (bool, error) {
	err := ctx.tx.Commit()
	return err == nil, err
}

type StorageAtomicOperator interface {
	plugin.Plugin

	CreateConfig(tx *sql.Tx, cfg *ConfigFile)

	ModifyConfig(tx *sql.Tx, cfg *ConfigFile)

	DeleteConfig(tx *sql.Tx, cfg *ConfigFile)

	FindOneConfig(tx *sql.Tx, query ConfigQuery) *ConfigFile

	CreateBetaConfig(tx *sql.Tx, cfg *ConfigBetaFile)

	ModifyBetaConfig(tx *sql.Tx, cfg *ConfigBetaFile)

	DeleteBetaConfig(tx *sql.Tx, cfg *ConfigBetaFile)

	FindOneBetaConfig(tx *sql.Tx, query ConfigQuery) *ConfigBetaFile

	CreateHistoryConfig(tx *sql.Tx, cfg *ConfigHistoryFile)

	DeleteHistoryConfig(tx *sql.Tx, cfg *ConfigHistoryFile)

	FindOneHistoryConfig(tx *sql.Tx, query ConfigQuery) *ConfigHistoryFile
}

type StorageOperator interface {
	PublishConfig(cfg *ConfigFile)

	SaveConfig(cfg *ConfigFile)

	DeleteConfig(cfg *ConfigFile)

	PublishBetaConfig(cfg *ConfigBetaFile)

	DeleteBetaConfig(cfg *ConfigFile)

	FindConfigs(query ConfigQuery) *[]ConfigFile
}

type EmbeddedRdsConfigStorage struct {
	rds storage.Rds
}

type ExternalRdsConfigStorage struct {
	rds *sql.DB
}
