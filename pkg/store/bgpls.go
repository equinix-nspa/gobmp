package store

import (
	"fmt"
	"sync"

	"github.com/sbezverk/gobmp/pkg/message"
)

// For nodes, key is [router-id+Asn+Lsid], as per sections 3.2.1.1 and 3.2.1.4 of RFC7752,
// We do not use name as part of the key since it is optional as per https://datatracker.ietf.org/doc/html/rfc7752#section-3.3.1.3
type nodeKey struct {
	IGPRouterId string
	Asn         uint32
	Lsid        uint32
}

// For links, key is [router-id, local-link IP, remote-link IP]
type linkKey struct {
	IGPRouterId  string
	LocalLinkIP  string
	RemoteLinkIP string
}

type BGPLSStore struct {
	// Read-write mutex to allow multiple readers
	mutex sync.RWMutex

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
	if link.IGPRouterID == "" || link.LocalLinkIP == "" || link.RemoteLinkIP == "" {
		return fmt.Errorf("empty string not expected in [%s,%s,%s] part of <%+v>", link.IGPRouterID, link.LocalLinkIP, link.RemoteLinkIP, link)
	}
	key := linkKey{
		IGPRouterId:  link.IGPRouterID,
		LocalLinkIP:  link.LocalLinkIP,
		RemoteLinkIP: link.RemoteLinkIP,
	}
	switch link.Action {
	case "add":
		s.links[key] = *link
	case "del":
		delete(s.links, key)
	default:
		return fmt.Errorf("unexpected action in %+v", link)
	}
	return nil
}

// Operation is in the node's's Action attribute
func (s *BGPLSStore) UpdateNode(node *message.LSNode) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check for empty values
	if node.IGPRouterID == "" || node.LSID == 0 || node.ASN == 0 {
		return fmt.Errorf("empty values not expected in <%+v>", node)
	}
	key := nodeKey{
		IGPRouterId: node.IGPRouterID,
		Asn:         node.ASN,
		Lsid:        node.LSID,
	}
	switch node.Action {
	case "add":
		s.nodes[key] = *node
	case "del":
		delete(s.nodes, key)
	default:
		return fmt.Errorf("unexpected action in %+v", node)
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

// The following is used when the caller has to transform what is in the store, it avoids
// double-traversal of the data which would happen with a call to Get() above followed by
// transformation of each entry
type GetLinkCB func(*message.LSLink)
type GetNodeCB func(*message.LSNode)

func (s *BGPLSStore) GetLinks(cb GetLinkCB) {
	for _, link := range s.links {
		cb(&link)
	}
}

func (s *BGPLSStore) GetNodes(cb GetNodeCB) {
	for _, node := range s.nodes {
		cb(&node)
	}
}

// New functions
func NewBGPLSStoreContents() *BGPLSStoreContents {
	return &BGPLSStoreContents{}
}

func NewBGPLSStore() *BGPLSStore {
	return &BGPLSStore{
		links: make(map[linkKey]message.LSLink),
		nodes: make(map[nodeKey]message.LSNode)}
}
