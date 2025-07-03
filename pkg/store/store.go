package store

import (
	"reflect"

	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
)

type Store struct {
	bgpls BGPLSStore
}

func (s *Store) GetBGPLS() *BGPLSStore {
	return &s.bgpls
}

func (s *Store) store(msg interface{}) {
	switch v := msg.(type) {
	case message.LSNode:
		if err := s.bgpls.UpdateNode(&v); err != nil {
			glog.Errorf("UpdateNode(%+v) failed:%+v", v, err)
		}
	case message.LSLink:
		if err := s.bgpls.UpdateLink(&v); err != nil {
			glog.Errorf("UpdateLink(%+v) failed:%+v", v, err)
		}
	default:
	}
}

func (s *Store) Store(msgQueue chan interface{}, stop chan struct{}) {
	for {
		select {
		case msg := <-msgQueue:
			s.store(msg)
		case <-stop:
			glog.Infof("Store() received interrupt, stopping.")
			return
		}
	}
}

func NewStore() *Store {
	return &Store{
		bgpls: *NewBGPLSStore(),
	}
}

// Returns true if v is a valid value and not zero/empty
func IsValidNonZero(v any) bool {
	return reflect.ValueOf(v).IsValid() && !reflect.ValueOf(v).IsZero()
}
