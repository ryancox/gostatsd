package main

import (
	"log"
	"net"

	"github.com/atlassian/gostatsd/statsd"
	"github.com/atlassian/gostatsd/types"
)

func main() {
	f := func(m *types.Metric) {
		log.Printf("%s", m)
	}
	r := statsd.MetricReceiver{
		SF: func() (net.PacketConn, error) {
			return net.ListenPacket("udp", ":8125")
		},
		Namespace:     "stats",
		MaxReaders:    1,
		MaxMessengers: 1,
		Handler:       statsd.HandlerFunc(f),
	}
	r.ListenAndReceive()

	select {}
}
