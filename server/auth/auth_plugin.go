// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"errors"
	"sync"
)

type OperationType int

var (
	AdminName = "admin"

	ReadOnly     OperationType = 0
	WriteOnly    OperationType = 1
	ReadAndWrite OperationType = 2

	ErrorTokenNotExist     = errors.New("token does not exist")
	ErrorTokenExpire       = errors.New("this token already expire")
	ErrorResourceNotFound  = errors.New("resource not found")
	ErrorForbid            = errors.New("forbid access this resource")
	ErrorOperateNotSupport = errors.New("this ops not support")
)

type User struct {
	id       int64
	username string
	password string
	role     string
}

func (u User) GetID() int64 {
	return u.id
}

func (u User) GetUsername() string {
	return u.username
}

func (u User) GetPassword() string {
	return u.password
}

func (u User) GetRole() string {
	return u.role
}

type Role struct {
	id          int64
	roleName    string
	permissions Permission
}

type Permission struct {
	resource  string
	operation OperationType
}

type Token struct {
	id    int64
	role  string
	token string
}

type SecurityCenter struct {
	ctx      context.Context
	rL       sync.RWMutex
	tL       sync.RWMutex
	roleMap  map[string]Role
	tokenMap map[string]Role
}

type RequestWrapper struct {
	Resource    string
	Op          OperationType
	AccessToken Token
}
