//  Copyright (c) 2020, Conf-Group. All rights reserved.
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.

package cluster

import (
	"github.com/Conf-Group/pole/transport"
)

type MemberCommunicate struct {
	membersClient map[string]*transport.RSocketClient
}


