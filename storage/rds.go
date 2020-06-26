// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package storage

import (
	"database/sql"
	"path/filepath"

	"nacos-go/sys"
	"nacos-go/utils"
)

type SqlRequest struct {
	sql  string
	args []interface{}
}

type QueryRequest struct {
	SqlRequest
}

type ModifyRequest struct {
	SqlRequest
}

type Rds struct {
	db *sql.DB
}

func CreateRds(cfg *sys.Config) *Rds {
	baseDir := filepath.Join(cfg.GetDataPath(), "conf.db")
	db, err := sql.Open("sqlite3", filepath.Join(baseDir, "sqlite3"))
	if err != nil {
		panic(err)
	}

	rds := &Rds{db: db}

	rds.initDB()

	return rds
}

func (r *Rds) initDB() {

}

func (r *Rds) ExecuteQuery(query QueryRequest) (results []map[string]interface{}, err error) {
	defer func() {
		if err := recover(); err != nil {
			results = nil
		}
	}()

	stmt := utils.Supplier(func() (i interface{}, err error) {
		return r.db.Prepare(query.sql)
	}).(*sql.Stmt)

	rows := utils.Supplier(func() (i interface{}, err error) {
		return stmt.Query(query.args)
	}).(*sql.Rows)

	columns, _ := rows.Columns()

	cache := make([]interface{}, len(columns))
	for index, _ := range cache {
		var placeholder interface{}
		cache[index] = &placeholder
	}

	for rows.Next() {
		utils.Runnable(func() error {
			return rows.Scan(cache)
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

func (r *Rds) ExecuteModify() error {

}
