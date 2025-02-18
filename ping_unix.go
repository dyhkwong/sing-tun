//go:build unix

package tun

import (
	"net"
	"net/netip"
	"os"
	"time"

	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/buf"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"

	"golang.org/x/sys/unix"
)

type uPinger struct {
	family byte
	conn   net.PacketConn
	rAddr  net.Addr
}

func NewPinger(destination netip.Addr) (Pinger, error) {
	var (
		family byte
		socket int
		err    error
	)
	if destination.Is4() {
		family = unix.AF_INET
		socket, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_ICMP)
	} else {
		family = unix.AF_INET6
		socket, err = unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_ICMPV6)
	}
	if err != nil {
		return nil, err
	}
	conn, err := net.FilePacketConn(os.NewFile(uintptr(socket), "ping"))
	if err != nil {
		return nil, E.Cause(err, "create ping conn")
	}
	return &uPinger{
		family: family,
		conn:   conn,
		rAddr:  M.SocksaddrFrom(destination, 0).UDPAddr(),
	}, nil
}

func (u *uPinger) WritePacket(packet *buf.Buffer) error {
	defer packet.Release()
	return common.Error(u.conn.WriteTo(packet.Bytes(), u.rAddr))
}

func (u *uPinger) ReadPacket() (*buf.Buffer, error) {
	packet := buf.NewPacket()
	_, _, err := packet.ReadPacketFrom(u.conn)
	if err != nil {
		packet.Release()
		return nil, err
	}
	return packet, nil
}

func (u *uPinger) SetReadDeadline(deadline time.Time) error {
	return u.conn.SetReadDeadline(deadline)
}

func (u *uPinger) Close() error {
	return u.conn.Close()
}
