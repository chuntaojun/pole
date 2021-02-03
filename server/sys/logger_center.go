// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	polerpc "github.com/pole-group/pole-rpc"
)

var (
	RaftKVLogger = polerpc.NewTestLogger("storage-raft-kv")

	CoreLogger = polerpc.NewTestLogger("pole-core")

	LookupLogger = polerpc.NewTestLogger("pole-cluster-lookup")

	DiscoveryLessLogger = polerpc.NewTestLogger("pole-discovery-less")
)
