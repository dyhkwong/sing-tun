package tun_test

import (
	"net/netip"
	"testing"

	"github.com/sagernet/sing-tun"
	"github.com/sagernet/sing-tun/internal/gtcpip/header"

	"github.com/stretchr/testify/require"
)

func TestPinger(t *testing.T) {
	dAddr := netip.MustParseAddr("223.5.5.5")
	pinger, err := tun.NewPinger(tun.PingOptions{
		Destination: dAddr,
	})
	require.NoError(t, err)
	message := tun.PingMessage(netip.Addr{}, dAddr, false, 1, 64, []byte("hello"))
	require.NotNil(t, message)
	require.NoError(t, pinger.WritePacket(message))
	packet, err := pinger.ReadPacket()
	require.NoError(t, err)
	ipHdr := header.IPv4(packet.Bytes())
	// why?
	ipHdr.SetTotalLength(uint16(len(ipHdr)))
	require.Equal(t, ipHdr.SourceAddr(), dAddr)
	icmpHdr := header.ICMPv4(ipHdr.Payload())
	require.Equal(t, header.ICMPv4EchoReply, icmpHdr.Type())
	require.Equal(t, uint16(1), icmpHdr.Sequence())
	require.Equal(t, "hello", string(icmpHdr.Payload()))
}
