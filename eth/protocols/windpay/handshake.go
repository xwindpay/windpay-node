package windpay

import (
	"github.com/ethereum/go-ethereum/p2p"
)

// SendWindpayCap sends the windpay capability message to the peer asynchronously.
// This is for backward compatibility with old nodes that expect a handshake.
// We send the message but don't wait for a response.
func (p *Peer) SendWindpayCap() {
	// Send capability message asynchronously for backward compatibility.
	// Old nodes expect this message to complete their handshake.
	// New nodes will ignore it via handleWindpayCap in handler.go.
	go func() {
		if err := p2p.Send(p.rw, WindpayCapMsg, &WindpayCapPacket{
			ProtocolVersion: p.version,
			Extra:           defaultExtra,
		}); err != nil {
			p.Log().Debug("Failed to send windpay capability message", "err", err)
		}
	}()
}
