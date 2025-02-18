package tun

import (
	"math/rand"
	"net/netip"
	"time"

	"github.com/sagernet/sing-tun/internal/gtcpip/checksum"
	"github.com/sagernet/sing-tun/internal/gtcpip/header"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/buf"
)

type Pinger interface {
	WritePacket(packet *buf.Buffer) error
	ReadPacket() (*buf.Buffer, error)
	SetReadDeadline(deadline time.Time) error
	Close() error
}

type PingOptions struct {
	Source           netip.Addr
	Destination      netip.Addr
	TTL              uint8
	TOS              uint8
	InterfaceMonitor DefaultInterfaceMonitor
}

func PingMessage(source netip.Addr, destination netip.Addr, headersIncluded bool, sequence uint16, ttl uint8, payload []byte) *buf.Buffer {
	var buffer *buf.Buffer
	if destination.Is4() {
		if headersIncluded {
			buffer = buf.NewSize(header.IPv4MinimumSize + header.ICMPv4MinimumSize + len(payload))
		} else {
			buffer = buf.NewSize(header.ICMPv4MinimumSize + len(payload))
		}
	} else {
		if headersIncluded {
			buffer = buf.NewSize(header.IPv6MinimumSize + header.ICMPv6MinimumSize + len(payload))
		} else {
			buffer = buf.NewSize(header.ICMPv6MinimumSize + len(payload))
		}
	}
	common.ClearArray(buffer.FreeBytes())
	if destination.Is4() {
		var ipHeader header.IPv4
		if headersIncluded {
			ipHeader = buffer.Extend(header.IPv4MinimumSize)
		}
		icmpHdr := header.ICMPv4(buffer.Extend(header.ICMPv4MinimumSize))
		buffer.Write(payload)
		if ipHeader != nil {
			ipHeader.Encode(&header.IPv4Fields{
				TotalLength: uint16(buffer.Len()),
				ID:          getPacketID(),
				TTL:         ttl,
				TOS:         0,
				Protocol:    uint8(header.ICMPv4ProtocolNumber),
				SrcAddr:     source,
				DstAddr:     destination,
			})
			ipHeader.SetChecksum(^ipHeader.CalculateChecksum())
		}
		icmpHdr.SetType(header.ICMPv4Echo)
		icmpHdr.SetSequence(sequence)
		icmpHdr.SetChecksum(header.ICMPv4Checksum(icmpHdr, checksum.Checksum(payload, 0)))
	} else {
		var ipHeader header.IPv6
		if headersIncluded {
			ipHeader = buffer.Extend(header.IPv6MinimumSize)
		}
		icmpHdr := header.ICMPv6(buffer.Extend(header.ICMPv6MinimumSize))
		buffer.Write(payload)
		if ipHeader != nil {
			ipHeader.Encode(&header.IPv6Fields{
				PayloadLength:     uint16(buffer.Len()),
				TransportProtocol: header.ICMPv6ProtocolNumber,
				HopLimit:          ttl,
				SrcAddr:           source,
				DstAddr:           destination,
			})
		}
		icmpHdr.SetType(header.ICMPv6EchoRequest)
		icmpHdr.SetSequence(sequence)
		icmpHdr.SetChecksum(header.ICMPv6Checksum(header.ICMPv6ChecksumParams{
			Header:      icmpHdr,
			Src:         source.AsSlice(),
			Dst:         destination.AsSlice(),
			PayloadCsum: checksum.Checksum(payload, 0),
			PayloadLen:  len(payload),
		}))
	}
	return buffer
}

func getPacketID() uint16 {
	packetID := uint16(rand.Uint32())
	if packetID == 0 {
		packetID = uint16(rand.Uint32())
	}
	return packetID
}
