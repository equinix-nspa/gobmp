package store_test

import (
	"testing"

	"github.com/sbezverk/gobmp/pkg/message"
	"github.com/sbezverk/gobmp/pkg/store"
	"github.com/stretchr/testify/require"
)

func checkStoreContentLengths(t *testing.T, s *store.BGPLSStore, numLinks int, numNodes int) {
	sc := s.Get()
	require.Equal(t, numNodes, len(sc.Nodes))
	require.Equal(t, numLinks, len(sc.Links))
}
func TestUpdateLinkErrorEmptyString(t *testing.T) {
	s := store.NewBGPLSStore()

	err := s.UpdateLink(&message.LSLink{Action: "add"})
	require.NotNil(t, err)

	checkStoreContentLengths(t, s, 0, 0)
}

func TestUpdateLinkErrorUnsupportedAction(t *testing.T) {
	s := store.NewBGPLSStore()

	err := s.UpdateLink(&message.LSLink{
		Action:      "xyz",
		IGPRouterID: "abcd",
		RouterIP:    "1.1.1.1",
		PeerIP:      "2.2.2.2"})
	require.NotNil(t, err)

	checkStoreContentLengths(t, s, 0, 0)
}

func TestUpdateNodeErrorEmptyString(t *testing.T) {
	s := store.NewBGPLSStore()

	err := s.UpdateNode(&message.LSNode{Action: "add"})
	require.NotNil(t, err)

	checkStoreContentLengths(t, s, 0, 0)
}

func TestUpdateNodeErrorUnsupportedAction(t *testing.T) {
	s := store.NewBGPLSStore()

	err := s.UpdateNode(&message.LSNode{Action: "",
		IGPRouterID: "abcd",
		Name:        "nodename"})
	require.NotNil(t, err)

	checkStoreContentLengths(t, s, 0, 0)
}

func TestLink(t *testing.T) {
	s := store.NewBGPLSStore()

	link := message.LSLink{
		Action:       "add",
		IGPRouterID:  "abcd",
		LocalLinkIP:  "1.1.1.1",
		RemoteLinkIP: "2.2.2.2"}
	err := s.UpdateLink(&link)
	require.Nil(t, err)
	// 1 link expected
	checkStoreContentLengths(t, s, 1, 0)

	// Add the same link
	err = s.UpdateLink(&link)
	require.Nil(t, err)
	// 1 link expected
	checkStoreContentLengths(t, s, 1, 0)

	// Now remove the link
	link.Action = "del"
	err = s.UpdateLink(&link)
	require.Nil(t, err)
	checkStoreContentLengths(t, s, 0, 0)
}

func TestNode(t *testing.T) {
	s := store.NewBGPLSStore()

	node := message.LSNode{
		Action:      "add",
		IGPRouterID: "abcd",
		Name:        "node"}
	err := s.UpdateNode(&node)
	require.Nil(t, err)
	// 1 node expected
	checkStoreContentLengths(t, s, 0, 1)

	// Add the same link
	err = s.UpdateNode(&node)
	require.Nil(t, err)
	// 1 node expected
	checkStoreContentLengths(t, s, 0, 1)

	// Now remove the node
	node.Action = "del"
	err = s.UpdateNode(&node)
	require.Nil(t, err)
	checkStoreContentLengths(t, s, 0, 0)
}
