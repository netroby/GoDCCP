// Copyright 2010 GoDCCP Authors. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package dccp

import (
	"os"
)

// inject() adds the packet h to the outgoing pipeline, without blocking.
// The pipeline is flushed continuously respecting the CongestionControl's
// rate-limiting policy.
// REMARK: inject() is called from inside a slk.Lock()
func (c *Conn) inject(h *Header) {
	panic("¿i?")
}