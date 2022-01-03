package pbcontroller

import (
	"sync"
	"github.com/upamanyus/hackv/pb"
)

type PBController struct {
	mu *sync.Mutex

	conf *pb.PBConfiguration
}

func (s *PBController) AddServerRPC() {
	// FIXME: impl
}


func MakePBController() {
	// FIXME: impl
}

func (s *PBController) Start() {
	// FIXME: impl
}
