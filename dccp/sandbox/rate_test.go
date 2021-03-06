// Copyright 2011-2013 GoDCCP Authors. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package sandbox

import (
	//"fmt"
	"testing"
	"github.com/petar/GoDCCP/dccp"
)

const (
	rateDuration           = 10e9   // Duration of rate test
	rateInterval           = 1e9
	ratePacketsPerInterval = 50
)

// TestRate tests whether a single connection's one-way client-to-server rate converges to
// limit imposed by connection in that the send rate has to:
//	(1) converge and stabilize, and
//	(2) the stable rate should 
//		(2.a) either be closely below the connection limit,
//		(2.b) or be closely above the connection limit (and maintain a drop rate below some threshold)
// A two-way test is not necessary as the congestion mechanisms in either direction are completely independent.
//
// NOTE: Pipe currently supports rate simulation in packets per time interval. If we want to test behavior
// under variable packet sizes, we need to implement rate simulation in bytes per interval.
func TestRate(t *testing.T) {

	env, _ := NewEnv("rate")
	clientConn, serverConn, clientToServer, _ := NewClientServerPipe(env)

	// Set rate limit on client-to-server connection
	clientToServer.SetWriteRate(rateInterval, ratePacketsPerInterval)

	cchan := make(chan int, 1)
	mtu := clientConn.GetMTU()
	buf := make([]byte, mtu)
	env.Go(func() {
		t0 := env.Now()
		for env.Now() - t0 < rateDuration {
			err := clientConn.Write(buf)
			if err != nil {
				t.Errorf("error writing (%s)", err)
				break
			}
		}
		// Close is necessary because otherwise, if no read timeout is in place, the
		// server sides hangs forever on Read
		clientConn.Close()
		close(cchan)
	}, "test client")

	schan := make(chan int, 1)
	env.Go(func() {
		for {
			_, err := serverConn.Read()
			if err == dccp.ErrEOF {
				break 
			} else if err != nil {
				t.Errorf("error reading (%s)", err)
				break
			}
		}
		serverConn.Close()
		close(schan)
	}, "test server")

	_, _ = <-cchan
	_, _ = <-schan

	clientConn.Abort()
	serverConn.Abort()

	env.NewGoJoin("end-of-test", clientConn.Joiner(), serverConn.Joiner()).Join()
	dccp.NewAmb("line", env).E(dccp.EventMatch, "Server and client done.")
	if err := env.Close(); err != nil {
		t.Errorf("error closing runtime (%s)", err)
	}
}
