// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/pojo"
)

func createTestLessor() *Lessor {
	return NewLessor(common.NewCtxPole())
}

func TestLessor_GrantLess(t *testing.T) {
	lessor := createTestLessor()

	instance := Instance{
		originInstance: &pojo.Instance{
			ServiceName:     "",
			Group:           "",
			Ip:              "127.0.0.1",
			Port:            80,
			ClusterName:     "",
			Weight:          0,
			Metadata:        nil,
			Ephemeral:       false,
			Enabled:         false,
			HealthCheckType: 0,
		},
	}

	lessor.GrantLess(instance)

	_, exist := lessor.lessRep[instance.GetKey()]
	assert.True(t, exist, "grant less must be success")
}
