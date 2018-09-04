package test

import (
	"context"
	"fmt"
	"testing"

	pstore "github.com/libp2p/go-libp2p-peerstore"
)

var peerstoreBenchmarks = map[string]func(pstore.Peerstore, chan *peerpair) func(*testing.B){
	"AddAddr":  benchmarkAddAddr,
	"AddAddrs": benchmarkAddAddrs,
	"SetAddrs": benchmarkSetAddrs,
	"GetAddrs": benchmarkGetAddrs,
}

func BenchmarkPeerstore(b *testing.B, factory PeerstoreFactory, variant string) {
	// Parameterises benchmarks to tackle peers with 1, 10, 100 multiaddrs.
	params := []struct {
		n  int
		ch chan *peerpair
	}{
		{1, make(chan *peerpair, 100)},
		{10, make(chan *peerpair, 100)},
		{100, make(chan *peerpair, 100)},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start all test peer producing goroutines, where each produces peers with as many
	// multiaddrs as the n field in the param struct.
	for _, p := range params {
		go addressProducer(ctx, b, p.ch, p.n)
	}

	for name, bench := range peerstoreBenchmarks {
		for _, p := range params {
			// Create a new peerstore.
			ps, closeFunc := factory()

			// Run the test.
			b.Run(fmt.Sprintf("%s-%dAddrs-%s", name, p.n, variant), bench(ps, p.ch))

			// Cleanup.
			if closeFunc != nil {
				closeFunc()
			}
		}
	}
}

func benchmarkAddAddr(ps pstore.Peerstore, addrs chan *peerpair) func(*testing.B) {
	return func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pp := <-addrs
			ps.AddAddr(pp.ID, pp.Addr[0], pstore.PermanentAddrTTL)
		}
	}
}

func benchmarkAddAddrs(ps pstore.Peerstore, addrs chan *peerpair) func(*testing.B) {
	return func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pp := <-addrs
			ps.AddAddrs(pp.ID, pp.Addr, pstore.PermanentAddrTTL)
		}
	}
}

func benchmarkSetAddrs(ps pstore.Peerstore, addrs chan *peerpair) func(*testing.B) {
	return func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pp := <-addrs
			ps.SetAddrs(pp.ID, pp.Addr, pstore.PermanentAddrTTL)
		}
	}
}

func benchmarkGetAddrs(ps pstore.Peerstore, addrs chan *peerpair) func(*testing.B) {
	return func(b *testing.B) {
		pp := <-addrs
		ps.SetAddrs(pp.ID, pp.Addr, pstore.PermanentAddrTTL)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ps.Addrs(pp.ID)
		}
	}
}
