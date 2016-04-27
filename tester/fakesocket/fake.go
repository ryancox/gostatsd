package fakesocket

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"
)

var FakeMetric = []byte("foo.bar.baz:2|c")

type FakeAddr struct{}

func (fa FakeAddr) Network() string { return "udp" }
func (fa FakeAddr) String() string  { return "127.0.0.1:8181" }

type FakePacketConn struct{}

func (fpc FakePacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n := copy(b, FakeMetric)
	return n, FakeAddr{}, nil
}
func (fpc FakePacketConn) WriteTo(b []byte, addr net.Addr) (int, error) { return 0, nil }
func (fpc FakePacketConn) Close() error                                 { return nil }
func (fpc FakePacketConn) LocalAddr() net.Addr                          { return FakeAddr{} }
func (fpc FakePacketConn) SetDeadline(t time.Time) error                { return nil }
func (fpc FakePacketConn) SetReadDeadline(t time.Time) error            { return nil }
func (fpc FakePacketConn) SetWriteDeadline(t time.Time) error           { return nil }

type FakeRandomPacketConn struct {
	FakePacketConn
}

func (frpc FakeRandomPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	num := rand.Int31n(10000) // Randomize metric name
	buf := new(bytes.Buffer)
	switch rand.Int31n(4) {
	case 0: // Counter
		fmt.Fprintf(buf, "statsd.tester.counter_%d:%f|c\n", num, rand.Float64()*100)
	case 1: // Gauge
		fmt.Fprintf(buf, "statsd.tester.gauge_%d:%f|g\n", num, rand.Float64()*100)
	case 2: // Timer
		n := rand.Intn(9) + 1
		for i := 0; i < n; i++ {
			fmt.Fprintf(buf, "statsd.tester.timer_%d:%f|ms\n", num, rand.Float64()*100)
		}
	case 3: // Set
		for i := 0; i < 100; i++ {
			fmt.Fprintf(buf, "statsd.tester.set_%d:%d|s\n", num, rand.Int31n(9)+1)
		}
	default:
		panic(errors.New("unreachable"))
	}
	n := copy(b, buf.Bytes())
	return n, FakeAddr{}, nil
}

func FakeSocketFactory() (net.PacketConn, error) {
	return FakeRandomPacketConn{}, nil
}
