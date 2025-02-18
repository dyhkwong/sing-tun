package tun

import (
	"context"
	"errors"
	"fmt"
	"github.com/sagernet/sing/common/atomic"
	"github.com/sagernet/sing/common/buf"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/pipe"
	"golang.org/x/sys/windows"
	"sync"
	"time"
	"unsafe"
)

type wPinger struct {
	options          PingOptions
	handle           windows.Handle
	recvChan         chan *buf.Buffer
	errChan          chan error
	readDeadline     pipe.Deadline
	readDeadlineTime atomic.TypedValue[time.Time]
	closeOnce        sync.Once
}

func NewPinger(options PingOptions) (Pinger, error) {
	var (
		handle windows.Handle
		err    error
	)
	if options.Destination.Is4() {
		handle, err = icmpCreateFile()
	} else {
		handle, err = icmp6CreateFile()
	}
	if err != nil {
		return nil, err
	}
	return &wPinger{
		options:      options,
		handle:       handle,
		readDeadline: pipe.MakeDeadline(),
		recvChan:     make(chan *buf.Buffer, 1),
		errChan:      make(chan error, 1),
	}, nil
}

func (w *wPinger) WritePacket(packet *buf.Buffer) error {
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(event)
	lAddr := w.options.Source.As4()
	rAddr := w.options.Destination.As4()
	response := buf.NewPacket()
	options := IPOptionInformation{
		Ttl: w.options.TTL,
		Tos: w.options.TOS,
	}
	err = icmpSendEcho2Ex(w.handle, event, 0, 0, &lAddr, &rAddr, packet.Bytes(), &options, response.FreeBytes(), 0)
	if !errors.Is(err, windows.ERROR_IO_PENDING) {
		return err
	}
	var timeout uint32
	if deadline := w.readDeadlineTime.Load(); !deadline.IsZero() {
		timeout = uint32(time.Until(deadline).Milliseconds())
	} else {
		timeout = windows.INFINITE
	}
	go func() {
		eventCode, err := windows.WaitForSingleObject(event, timeout)
		if err == nil && eventCode == windows.WAIT_OBJECT_0 {
			w.parseReplies(response)
		} else {
			if err == nil {
				err = E.New("wait failed, code: ", fmt.Sprintf("%x", eventCode))
			}
			packet.Release()
			select {
			case w.errChan <- err:
			default:
			}
		}
	}()
	return nil
}

func (w *wPinger) parseReplies(buffer *buf.Buffer) {
	replyCount, err := icmpParseReplies(buffer.FreeBytes())
	if err != nil {
		buffer.Release()
		w.errChan <- err
		return
	}
	replies := unsafe.Slice((*IcmpEchoReply)(unsafe.Pointer(&buffer.Index(0)[0])), replyCount)
	for _, reply := range replies {

	}
}

func (w *wPinger) ReadPacket() (*buf.Buffer, error) {
	select {
	case packet := <-w.recvChan:
		return packet, nil
	case <-w.readDeadline.Wait():
		return nil, context.DeadlineExceeded
	case err := <-w.errChan:
		return nil, err
	}
}

func (w *wPinger) SetReadDeadline(deadline time.Time) error {
	w.readDeadline.Set(deadline)
	w.readDeadlineTime.Store(deadline)
	return nil
}

func (w *wPinger) Close() error {
	var err error
	w.closeOnce.Do(func() {
		err = windows.CloseHandle(w.handle)
	})
	return err
}

type IPOptionInformation struct {
	Ttl         byte
	Tos         byte
	Flags       byte
	OptionsSize byte
	OptionsData *byte
}

type IcmpEchoReply struct {
	Address       [4]byte
	Status        uint32
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          unsafe.Pointer
	Options       IPOptionInformation
}
