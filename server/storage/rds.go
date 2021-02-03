// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/mattn/go-sqlite3" // nolint:golint

	"github.com/pole-group/pole/server/sys"
	"github.com/pole-group/pole/utils"
)

var (
	ErrorNotRegister = errors.New("this table is not registered for automatic conversion")

	MapperCache = make(map[string]func(result map[string]interface{}) interface{})
)

func RegisterMapper(table string, mapper func(result map[string]interface{}) interface{}) {
	MapperCache[table] = mapper
}

func RowMap(table string, result map[string]interface{}) interface{} {
	if mapper, isOk := MapperCache[table]; isOk {
		return mapper(result)
	}
	panic(ErrorNotRegister)
}

type SQLRequest struct {
	SQL  string
	Args []interface{}
}

type QueryRequest struct {
	SQLRequest
}

type ModifyRequest struct {
	SQLRequest
	No int
}

type ModifyReqs []ModifyRequest

// Len()
func (s ModifyReqs) Len() int {
	return len(s)
}

// Less()
func (s ModifyReqs) Less(i, j int) bool {
	return s[i].No < s[j].No
}

// Swap()
func (s ModifyReqs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type Rds interface {
	DB() *sql.DB

	initDB(f func(db *sql.DB))

	ExecuteQuery(ctx context.Context, query QueryRequest) (results []map[string]interface{}, err error)

	ExecuteModify(ctx context.Context, reqs ModifyReqs) (err error)

	Transaction(ctx context.Context, f func(db *sql.DB) error) error
}

func CreateRds(cfg *sys.PoleConfig, initialize func(db *sql.DB)) (Rds, error) {
	if cfg.IsEmbedded {
		utils.MkdirAllIfNotExist(cfg.GetDataPath(), os.ModePerm)
		db, err := sql.Open("sqlite3", filepath.Join(cfg.GetDataPath(), "conf.db"))
		if err != nil {
			return nil, err
		}
		rds := &EmbeddedRDS{db: db}
		rds.initDB(initialize)
		return rds, nil
	}
	return nil, fmt.Errorf("unsupport!")
}

type EmbeddedRDS struct {
	db *sql.DB
}

func (r *EmbeddedRDS) DB() *sql.DB {
	return r.db
}

func (r *EmbeddedRDS) initDB(f func(db *sql.DB)) {
	f(r.db)
}

func (r *EmbeddedRDS) ExecuteQuery(ctx context.Context, query QueryRequest) (results []map[string]interface{}, err error) {
	defer func() {
		if err := recover(); err != nil {
			results = nil
		}
	}()

	rows := utils.Supplier(func() (i interface{}, err error) {
		return r.db.Query(query.SQL, query.Args...)
	}).(*sql.Rows)

	columns, _ := rows.Columns()

	cache := make([]interface{}, len(columns))
	for index := range cache {
		var placeholder interface{}
		cache[index] = &placeholder
	}

	for rows.Next() {
		utils.Runnable(func() error {
			return rows.Scan(cache...)
		})

		record := make(map[string]interface{})
		for i, d := range cache {
			record[columns[i]] = d
		}

		results = append(results, record)
	}

	err = nil

	return
}

func (r *EmbeddedRDS) ExecuteModify(ctx context.Context, reqs ModifyReqs) (err error) {
	sort.Sort(reqs)
	db := r.db

	tx, err := db.Begin()
	if err != nil {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			err = tx.Rollback()
			if err != nil {
				fmt.Printf("db transaction rollback has error : %s\n", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				fmt.Printf("db transaction commit has error : %s\n", err)
			}
		}
	}()

	for _, req := range reqs {
		_ = utils.Supplier(func() (i interface{}, err error) {
			return tx.Exec(req.SQL, req.Args...)
		}).(sql.Result)
	}

	return
}

func (r *EmbeddedRDS) Transaction(ctx context.Context, f func(db *sql.DB) error) error {
	return transaction(ctx, r.db, f)
}

func transaction(ctx context.Context, db *sql.DB, f func(db *sql.DB) error) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})

	if err != nil {
		return err
	}

	err = f(db)
	if err != nil {
		return tx.Rollback()
	}
	return tx.Commit()
}
