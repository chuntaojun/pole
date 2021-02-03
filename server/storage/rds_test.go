//  Copyright (c) 2020, pole-group. All rights reserved.
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	
	"github.com/stretchr/testify/assert"

	"github.com/pole-group/pole/server/sys"
	"github.com/pole-group/pole/utils"
)

type User struct {
	ID       int64
	Username string
	Password string
}

func Test_RdsQuery(t *testing.T) {
	_ = os.RemoveAll(utils.GetStringFromEnv("HOME"))
	
	RegisterMapper("users", func(result map[string]interface{}) interface{} {
		u := User{}
		u.ID = (*result["id"].(*interface{})).(int64)
		u.Username = (*result["username"].(*interface{})).(string)
		u.Password = (*result["password"].(*interface{})).(string)
		return u
	})
	
	rds, _ := CreateRds(&sys.PoleConfig{
		BaseDir: filepath.Join(utils.GetStringFromEnv("HOME"), "conf_rds"),
	}, func(db *sql.DB) {
	
	})
	
	_, err := rds.DB().Exec(
		`create table if not exists users(
    id integer primary key autoincrement ,
    username varchar(255) not null ,
    password varchar(255) not null
)`)
	
	if err != nil {
		t.Error(err)
		return
	}
	
	_, err = rds.DB().Exec(`insert into users(username, password) values (?, ?)`, "liaochuntao",
		"liaochuntao")
	
	if err != nil {
		t.Error(err)
		return
	}
	
	results, err := rds.ExecuteQuery(context.TODO(), QueryRequest{
		SQLRequest{
			SQL:  "select * from users",
			Args: nil,
		},
	})
	
	if err != nil {
		t.Error(err)
	}
	
	var users []User
	
	for _, m := range results {
		users = append(users, RowMap("users", m).(User))
	}
	assert.EqualValues(t, 1, len(users))
	assert.NotNil(t, users[0])
	assert.EqualValues(t, "liaochuntao", users[0].Username)
	fmt.Printf("users : %+v\n", users)
}
