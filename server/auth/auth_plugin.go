// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Conf-Group/pole/constants"
	"github.com/Conf-Group/pole/utils"
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

func CreateSecurityCenter(ctx context.Context) *SecurityCenter {
	return &SecurityCenter{
		ctx:      ctx,
		roleMap:  make(map[string]Role),
		tokenMap: make(map[string]Role),
	}
}

func (s *SecurityCenter) startRoleRefresh() {
	utils.DoTickerSchedule(s.ctx, func() {

	}, time.Duration(30)*time.Second)
}

func (s *SecurityCenter) startTokenRefresh() {
	utils.DoTickerSchedule(s.ctx, func() {

	}, time.Duration(30)*time.Second)
}

func (s *SecurityCenter) HasPermission(metadata map[string]string, op OperationType) (bool, error) {
	token := metadata[constants.TokenKey]
	resource := s.buildResource(metadata)
	// Judge whether token exists
	if role, exist := s.tokenMap[token]; exist {
		return s.authFilter(resource, role, op)
	}
	return false, ErrorTokenNotExist
}

func (s *SecurityCenter) authFilter(resource string, role Role, op OperationType) (bool, error) {
	defer func() {
		s.tL.RUnlock()
	}()
	s.tL.RLock()

	if strings.Compare(AdminName, role.roleName) == 0 {
		return true, nil
	}

	// what permissions do I have to find this resource
	p := role.permissions
	reg := strings.ReplaceAll(p.resource, "*", ".*")
	isOk, err := regexp.Match(reg, []byte(resource))
	if err != nil {
		return false, err
	}

	if !isOk {
		return false, ErrorForbid
	}

	if op == p.operation {
		return true, nil
	}

	return false, fmt.Errorf("forbid %s this resource, it just allow %s", parseOperationName(op),
		parseOperationName(p.operation))
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
		panic(ErrorOperateNotSupport)
	}
}

// {namespace}:{group}
func (s *SecurityCenter) buildResource(header map[string]string) string {
	namespaceID := header[constants.NamespaceID]
	group := header[constants.Group]
	resource := ""
	if namespaceID != "" {
		resource += namespaceID
	}
	resource += ":"

	if group == "" {
		resource += "*"
	} else {
		resource += group
	}
	return resource
}
