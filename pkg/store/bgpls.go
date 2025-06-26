package store

import (
	"fmt"
	"sync"

	"github.com/sbezverk/gobmp/pkg/message"
)

// For nodes, key is [router-id+name]
type nodeKey struct {
	IGPRouterId string
	Name        string
}

// For links, key is [router-id, local IP, remote-IP]
type linkKey struct {
	IGPRouterId string
	RouterIP    string
	PeerIP      string
}

type BGPLSStore struct {
	// Read-write mutex to allow multiple readers
	mutex sync.RWMutex

	// What is in the store
	//state BGPLSStoreContents
	// BGP-LS nodes
	nodes map[nodeKey]message.LSNode
	// BGP-LS links
	links map[linkKey]message.LSLink
	// No support for prefixes yet
}

// Contents we return via Get()
type BGPLSStoreContents struct {
	// BGP-LS nodes
	Nodes []message.LSNode
	// BGP-LS links
	Links []message.LSLink
	// No support for prefixes yet
}

// Operation is in the link's Action attribute
func (s *BGPLSStore) UpdateLink(link *message.LSLink) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check for empty strings
	if link.IGPRouterID == "" || link.RouterIP == "" || link.PeerIP == "" {
		return fmt.Errorf("Empty string not expected in %+v", link)
	}
	key := linkKey{
		IGPRouterId: link.IGPRouterID,
		RouterIP:    link.RouterIP,
		PeerIP:      link.PeerIP,
	}
	switch link.Action {
	case "add":
		s.links[key] = *link
	case "del":
		delete(s.links, key)
	default:
		return fmt.Errorf("Unexpected action in %+v", link)
	}
	return nil
}

// Operation is in the node's's Action attribute
func (s *BGPLSStore) UpdateNode(node *message.LSNode) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check for empty strings
	if node.IGPRouterID == "" || node.Name == "" {
		return fmt.Errorf("Empty string not expected in %+v", node)
	}
	key := nodeKey{
		IGPRouterId: node.IGPRouterID,
		Name:        node.Name,
	}
	switch node.Action {
	case "add":
		s.nodes[key] = *node
	case "del":
		delete(s.nodes, key)
	default:
		return fmt.Errorf("Unexpected action in %+v", node)
	}
	return nil
}

func (s *BGPLSStore) Get() *BGPLSStoreContents {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	contents := NewBGPLSStoreContents()
	for _, value := range s.links {
		contents.Links = append(contents.Links, value)
	}
	for _, value := range s.nodes {
		contents.Nodes = append(contents.Nodes, value)
	}

	return contents
}

func NewBGPLSStoreContents() *BGPLSStoreContents {
	return &BGPLSStoreContents{}
}

func NewBGPLSStore() *BGPLSStore {
	return &BGPLSStore{
		links: make(map[linkKey]message.LSLink),
		nodes: make(map[nodeKey]message.LSNode)}
}
