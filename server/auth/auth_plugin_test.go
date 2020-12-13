// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Conf-Group/pole/constants"
)

var (
	NamespaceOne   = "namespace_one"
	NamespaceTwo   = "namespace_two"
	NamespaceThree = "namespace_three"

	TokenOne   = "token_one"
	TokenTwo   = "token_two"
	TokenThree = "token_three"
)

func Test_AuthFilter(t *testing.T) {
	ctx := context.Background()
	s := CreateSecurityCenter(ctx)
	initResource(s)

	var isOk bool
	var err error

	// admin role test start
	// 管理员任何操作都可以
	adminHeader := map[string]string{
		constants.TokenKey:    TokenOne,
		constants.NamespaceID: NamespaceOne,
	}

	isOk, err = s.HasPermission(adminHeader, ReadOnly)
	assert.Nil(t, err, "must not occur error")
	assert.True(t, isOk, "must be ok")

	isOk, err = s.HasPermission(adminHeader, WriteOnly)
	assert.Nil(t, err, "must not occur error")
	assert.True(t, isOk, "must be ok")

	isOk, err = s.HasPermission(adminHeader, ReadAndWrite)
	assert.Nil(t, err, "must not occur error")
	assert.True(t, isOk, "must be ok")

	// admin role test end

	// developer role test start
	developHeader := map[string]string{
		constants.TokenKey:    TokenTwo,
		constants.NamespaceID: NamespaceTwo,
	}

	// 开发者角色只允许写操作
	isOk, err = s.HasPermission(developHeader, WriteOnly)
	assert.Nil(t, err, "must not occur error")
	assert.True(t, isOk, "must be true")

	// 开发者角色读操作失败
	isOk, err = s.HasPermission(developHeader, ReadOnly)
	assert.EqualValues(t, errors.New("forbid read this resource, it just allow write"), err)
	assert.False(t, isOk, "must be false")

	// 开发者角色读写操作失败
	isOk, err = s.HasPermission(developHeader, ReadAndWrite)
	assert.EqualValues(t, errors.New("forbid read and write this resource, it just allow write"), err)
	assert.False(t, isOk, "must be false")

	// token 访问错误的命名空间
	developHeader = map[string]string{
		constants.TokenKey:    TokenTwo,
		constants.NamespaceID: NamespaceOne,
	}

	isOk, err = s.HasPermission(developHeader, WriteOnly)
	assert.EqualValues(t, ErrorForbid, err)
	assert.False(t, isOk, "must be false")

	// developer role test end

	// tester role test start
	// 测试者正常的操作，正确的token以及访问正确的命名空间
	testHeader := map[string]string{
		constants.TokenKey:    TokenThree,
		constants.NamespaceID: NamespaceThree,
	}
	isOk, err = s.HasPermission(testHeader, ReadOnly)
	assert.Nil(t, err, "must no occur error")
	assert.True(t, isOk, "must be true")

	// 使用了不存在的 Token 进行资源操作
	testHeader = map[string]string{
		constants.TokenKey:    "TokenFour",
		constants.NamespaceID: NamespaceOne,
	}

	isOk, err = s.HasPermission(testHeader, ReadOnly)
	assert.Equal(t, ErrorTokenNotExist, err)
	assert.False(t, isOk, "must be false")

	// tester role test end
}

func initResource(center *SecurityCenter) {
	rAdmin := Role{
		id: 1,
		permissions: Permission{
			resource:  "*:*",
			operation: ReadAndWrite,
		},
		roleName: "admin",
	}

	rDevelop := Role{
		id: 2,
		permissions: Permission{
			resource:  NamespaceTwo + ":*",
			operation: WriteOnly,
		},
		roleName: "develop",
	}

	rTest := Role{
		id: 3,
		permissions: Permission{
			resource:  NamespaceThree + ":*",
			operation: ReadOnly,
		},
		roleName: "test",
	}

	center.roleMap[rAdmin.roleName] = rAdmin
	center.roleMap[rDevelop.roleName] = rDevelop
	center.roleMap[rTest.roleName] = rTest

	center.tokenMap[TokenOne] = rAdmin
	center.tokenMap[TokenTwo] = rDevelop
	center.tokenMap[TokenThree] = rTest
}
