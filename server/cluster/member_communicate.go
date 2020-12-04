package cluster

import (
	"github.com/Conf-Group/pole/transport/rsocket"
)

type MemberCommunicate struct {
	membersClient map[string]*rsocket.RSocketClient
}


