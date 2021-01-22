//  Copyright (c) 2020, pole-group. All rights reserved.
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.

package cluster

import (
	pole_rpc "github.com/pole-group/pole-rpc"
)

type MemberCommunicate struct {
	membersClient map[string]*pole_rpc.TransportClient
}


