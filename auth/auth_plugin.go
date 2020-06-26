// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	
	"nacos-go/utils"
)

type OperationType int

var (
	ReadOnly     OperationType = 0
	WriteOnly    OperationType = 1
	ReadAndWrite OperationType = 2
	
	TokenNotExist = errors.New("token does not exist")
)

type User struct {
	id       int64
	username string
	password string
	role     string
}

func (u User) GetId() int64 {
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
	roleMap  map[string]Role
	tL       sync.RWMutex
	tokenMap map[string]Role
}

func CreateSecurityCenter(ctx context.Context) *SecurityCenter {
	return &SecurityCenter{
		ctx: ctx,
	}
}

func (s *SecurityCenter) startRoleRefresh() {
	utils.DoTickerSchedule(func() {
	
	}, time.Duration(30)*time.Second, s.ctx)
}

func (s *SecurityCenter) startTokenRefresh() {
	utils.DoTickerSchedule(func() {
	
	}, time.Duration(30)*time.Second, s.ctx)
}

func (s *SecurityCenter) Filter(header map[string]string) (bool, error) {
	token := header["Token"]
	resource := header["Resource"]
	if role, exist := s.tokenMap[token]; exist {
		return s.authFilter(token, resource, role.permissions.operation)
	} else {
		return false, TokenNotExist
	}
}

func (s *SecurityCenter) authFilter(token, resource string, operation OperationType) (bool, error) {
	defer func() {
		s.tL.RUnlock()
	}()
	
	s.tL.RLock()
	
	v, exist := s.tokenMap[token]
	if !exist {
		return false, fmt.Errorf("this token alreay expire")
	}
	
	p := v.permissions
	if strings.Compare(p.resource, resource) != 0 {
		return false, fmt.Errorf("forbiden access this resource")
	}
	
	if operation == p.operation {
		return true, nil
	}
	
	return false, fmt.Errorf("forbiden operation this resource, it just allow %s", parseOperationName(p.operation))
	
}

func parseOperationName(ops OperationType) string {
	switch ops {
	case ReadOnly:
		return "read"
	case WriteOnly:
		return "write"
	case ReadAndWrite:
		return "read and write"
	default:
		panic("this ops not support")
	}
}

func (s *SecurityCenter) analyzeResource(resource string) map[string]string {
	detail := strings.Split(resource, "@@")
	info := make(map[string]string)
	info["namespace"] = detail[0]
	info["group"] = detail[1]
	return info
}
