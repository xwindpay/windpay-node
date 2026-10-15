package eth

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/protocols/windpay"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// windpayHandler implements the windpay.Backend interface to handle the various network
// packets that are sent as broadcasts.
type windpayHandler handler

func (h *windpayHandler) Chain() *core.BlockChain { return h.chain }

// RunPeer is invoked when a peer joins on the `windpay` protocol.
func (h *windpayHandler) RunPeer(peer *windpay.Peer, hand windpay.Handler) error {
	// Send capability message asynchronously for backward compatibility.
	// Old nodes expect this message to complete their handshake.
	// We don't wait for response - just send and continue.
	peer.SendWindpayCap()

	return (*handler)(h).runWindpayExtension(peer, hand)
}

// PeerInfo retrieves all known `windpay` information about a peer.
func (h *windpayHandler) PeerInfo(id enode.ID) interface{} {
	if p := h.peers.peer(id.String()); p != nil && p.windpayExt != nil {
		return p.windpayExt.info()
	}
	return nil
}

// Handle is invoked from a peer's message handler when it receives a new remote
// message that the handler couldn't consume and serve itself.
func (h *windpayHandler) Handle(peer *windpay.Peer, packet windpay.Packet) error {
	// DeliverSnapPacket is invoked from a peer's message handler when it transmits a
	// data packet for the local node to consume.
	switch packet := packet.(type) {
	case *windpay.VotesPacket:
		return h.handleVotesBroadcast(peer, packet.Votes)

	default:
		return fmt.Errorf("unexpected windpay packet type: %T", packet)
	}
}

// handleVotesBroadcast is invoked from a peer's message handler when it transmits a
// votes broadcast for the local node to process. Per-peer rate limiting happens
// upstream in windpay.handleVotes via IsOverLimitAfterReceivingVotes.
func (h *windpayHandler) handleVotesBroadcast(peer *windpay.Peer, votes []*types.VoteEnvelope) error {
	// Here we only put the first vote, to avoid ddos attack by sending a large batch of votes.
	// This won't abandon any valid vote, because one vote is sent every time referring to func voteBroadcastLoop
	if len(votes) > 0 {
		h.votepool.PutVote(votes[0])
	}

	return nil
}
