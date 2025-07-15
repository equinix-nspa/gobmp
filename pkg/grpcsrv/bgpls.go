package grpcsrv

import (
	"github.com/sbezverk/gobmp/pkg/api/generated"
	"github.com/sbezverk/gobmp/pkg/message"
	"github.com/sbezverk/gobmp/pkg/store"
)

func GetBGPLS(bgplsStore *store.BGPLSStore) *generated.GetLSResponse {
	response := &generated.GetLSResponse{}

	addLinkCB := func(msg *message.LSLink) {
		pbLink := &generated.LSLink{}

		// Action not needed here since we're maintaining state
		if store.IsValidNonZero(msg.RouterHash) {
			pbLink.RouterHash = msg.RouterHash
		}
		if store.IsValidNonZero(msg.RouterIP) {
			pbLink.RouterIp = msg.RouterIP
		}
		if store.IsValidNonZero(msg.DomainID) {
			pbLink.DomainId = msg.DomainID
		}
		if store.IsValidNonZero(msg.PeerHash) {
			pbLink.PeerHash = msg.PeerHash
		}
		if store.IsValidNonZero(msg.PeerIP) {
			pbLink.PeerIp = msg.PeerIP
		}
		if store.IsValidNonZero(msg.PeerType) {
			pbLink.PeerType = uint32(msg.PeerType)
		}
		if store.IsValidNonZero(msg.PeerASN) {
			pbLink.PeerAsn = msg.PeerASN
		}
		if store.IsValidNonZero(msg.Timestamp) {
			pbLink.Timestamp = msg.Timestamp
		}
		if store.IsValidNonZero(msg.IGPRouterID) {
			pbLink.IgpRouterId = msg.IGPRouterID
		}
		if store.IsValidNonZero(msg.Protocol) {
			pbLink.Protocol = msg.Protocol
		}
		if store.IsValidNonZero(msg.ProtocolID) {
			pbLink.ProtocolId = uint32(msg.ProtocolID)
		}
		if store.IsValidNonZero(msg.AreaID) {
			pbLink.AreaId = msg.AreaID
		}
		if store.IsValidNonZero(msg.Nexthop) {
			pbLink.NextHop = msg.Nexthop
		}
		if store.IsValidNonZero(msg.LocalLinkIP) {
			pbLink.LocalLinkIp = msg.LocalLinkIP
		}
		if store.IsValidNonZero(msg.RemoteLinkIP) {
			pbLink.RemoteLinkIp = msg.RemoteLinkIP
		}
		if store.IsValidNonZero(msg.LocalLinkID) {
			pbLink.LocalLinkId = msg.LocalLinkID
		}
		if store.IsValidNonZero(msg.RemoteLinkID) {
			pbLink.RemoteLinkId = msg.RemoteLinkID
		}
		if store.IsValidNonZero(msg.IGPMetric) {
			pbLink.IgpMetric = msg.IGPMetric
		}
		if store.IsValidNonZero(msg.TEDefaultMetric) {
			pbLink.TeDefaultMetric = msg.TEDefaultMetric
		}
		if store.IsValidNonZero(msg.RemoteNodeHash) {
			pbLink.RemoteNodeHash = msg.RemoteNodeHash
		}
		if store.IsValidNonZero(msg.LocalNodeHash) {
			pbLink.LocalNodeHash = msg.LocalNodeHash
		}
		if store.IsValidNonZero(msg.RemoteIGPRouterID) {
			pbLink.RemoteIgpRouterId = msg.RemoteIGPRouterID
		}
		if store.IsValidNonZero(msg.LocalNodeASN) {
			pbLink.LocalNodeAsn = msg.LocalNodeASN
		}
		if store.IsValidNonZero(msg.RemoteNodeASN) {
			pbLink.RemoteNodeAsn = msg.RemoteNodeASN
		}
		if store.IsValidNonZero(msg.LinkName) {
			pbLink.LinkName = msg.LinkName
		}
		if store.IsValidNonZero(msg.AdminGroup) {
			// TBD rename to admin_group
			pbLink.AdminGrpup = msg.AdminGroup
		}
		if store.IsValidNonZero(msg.MaxLinkBW) {
			pbLink.MaxLinkBw = msg.MaxLinkBW
		}
		if store.IsValidNonZero(msg.MaxResvBW) {
			pbLink.MaxResvBw = msg.MaxResvBW
		}
		if store.IsValidNonZero(msg.LinkProtection) {
			pbLink.LinkProtection = uint32(msg.LinkProtection)
		}
		if store.IsValidNonZero(msg.MPLSProtoMask) {
			pbLink.MplsProtoMask = uint32(msg.MPLSProtoMask)
		}
		if store.IsValidNonZero(msg.BGPRouterID) {
			pbLink.BgpRouterId = msg.BGPRouterID
		}
		if store.IsValidNonZero(msg.BGPRemoteRouterID) {
			pbLink.BgpRemoteRouterId = msg.BGPRemoteRouterID
		}
		if store.IsValidNonZero(msg.MemberAS) {
			pbLink.MemberAs = msg.MemberAS
		}
		if store.IsValidNonZero(msg.UnidirLinkDelay) {
			pbLink.UnidirLinkDelay = msg.UnidirLinkDelay
		}
		if store.IsValidNonZero(msg.UnidirLinkDelayMinMax) {
			pbLink.UnidirLinkDelayMinMax = msg.UnidirLinkDelayMinMax
		}
		if store.IsValidNonZero(msg.UnidirDelayVariation) {
			pbLink.UnidirLinkDelayVariation = msg.UnidirDelayVariation
		}
		if store.IsValidNonZero(msg.UnidirPacketLoss) {
			pbLink.UnidirPacketLoss = msg.UnidirPacketLoss
		}
		if store.IsValidNonZero(msg.UnidirResidualBW) {
			pbLink.UnidirResidualBw = msg.UnidirResidualBW
		}
		if store.IsValidNonZero(msg.UnidirAvailableBW) {
			pbLink.UnidirAvailableBw = msg.UnidirAvailableBW
		}
		if store.IsValidNonZero(msg.UnidirBWUtilization) {
			pbLink.UnidirBwUtilization = msg.UnidirBWUtilization
		}
		if store.IsValidNonZero(msg.IsAdjRIBInPost) {
			pbLink.IsAdjRibInPost = msg.IsAdjRIBInPost
		}
		if store.IsValidNonZero(msg.IsAdjRIBOutPost) {
			pbLink.IsAdjRibOutPost = msg.IsAdjRIBOutPost
		}
		if store.IsValidNonZero(msg.IsLocRIBFiltered) {
			pbLink.IsLocalRibFiltered = msg.IsLocRIBFiltered
		}
		response.Links = append(response.Links, pbLink)
	}
	addNodeCB := func(msg *message.LSNode) {
		pbNode := &generated.LSNode{}

		if store.IsValidNonZero(msg.Key) {
			pbNode.Key = msg.Key
		}
		if store.IsValidNonZero(msg.ID) {
			pbNode.Id = msg.ID
		}
		if store.IsValidNonZero(msg.Rev) {
			pbNode.Rev = msg.Rev
		}
		// Action not needed here since we're maintaining state

		if store.IsValidNonZero(msg.RouterHash) {
			pbNode.RouterHash = msg.RouterHash
		}
		if store.IsValidNonZero(msg.DomainID) {
			pbNode.DomainId = msg.DomainID
		}
		if store.IsValidNonZero(msg.RouterIP) {
			pbNode.RouterIp = msg.RouterIP
		}
		if store.IsValidNonZero(msg.PeerHash) {
			pbNode.PeerHash = msg.PeerHash
		}
		if store.IsValidNonZero(msg.PeerIP) {
			pbNode.PeerIp = msg.PeerIP
		}
		if store.IsValidNonZero(msg.PeerType) {
			pbNode.PeerType = uint32(msg.PeerType)
		}
		if store.IsValidNonZero(msg.PeerASN) {
			pbNode.PeerAsn = msg.PeerASN
		}
		if store.IsValidNonZero(msg.Timestamp) {
			pbNode.Timestamp = msg.Timestamp
		}
		if store.IsValidNonZero(msg.IGPRouterID) {
			pbNode.IgpRouterId = msg.IGPRouterID
		}
		if store.IsValidNonZero(msg.RouterID) {
			pbNode.RouterId = msg.RouterID
		}
		if store.IsValidNonZero(msg.ASN) {
			pbNode.Asn = msg.ASN
		}
		if store.IsValidNonZero(msg.AreaID) {
			pbNode.AreaId = msg.AreaID
		}
		if store.IsValidNonZero(msg.Protocol) {
			pbNode.Protocol = msg.Protocol
		}
		if store.IsValidNonZero(msg.ProtocolID) {
			pbNode.ProtocolId = uint32(msg.ProtocolID)
		}
		if store.IsValidNonZero(msg.NodeFlags) {
			pbNode.NodeFlags = &generated.LSNodeAttrFlags{}
			pbNode.NodeFlags.OFlag = msg.NodeFlags.OFlag
			pbNode.NodeFlags.TFlag = msg.NodeFlags.TFlag
			pbNode.NodeFlags.EFlag = msg.NodeFlags.EFlag
			pbNode.NodeFlags.BFlag = msg.NodeFlags.BFlag
			pbNode.NodeFlags.RFlag = msg.NodeFlags.RFlag
			// TBD change proto to have VFlag and re-generate code from proto
			pbNode.NodeFlags.FFlag = msg.NodeFlags.VFlag
		}
		if store.IsValidNonZero(msg.Name) {
			pbNode.Name = msg.Name
		}
		if store.IsValidNonZero(msg.IsAdjRIBInPost) {
			pbNode.IsAdjRibInPost = msg.IsAdjRIBInPost
		}
		if store.IsValidNonZero(msg.IsAdjRIBOutPost) {
			pbNode.IsAdjRibOutPost = msg.IsAdjRIBOutPost
		}
		if store.IsValidNonZero(msg.IsLocRIBFiltered) {
			pbNode.IsLocalRibFiltered = msg.IsLocRIBFiltered
		}
		response.Nodes = append(response.Nodes, pbNode)
	}
	bgplsStore.GetLinks(addLinkCB)
	bgplsStore.GetNodes(addNodeCB)
	return response
}
