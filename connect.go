//
// Copyright © 2011-2012 Guy M. Allard
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package stompngo

import (
	"bufio"
	"net"
	"strings"
)

// Primary STOMP Connect.  For STOMP 1.1+ the Headers parameter must contain
// those headers required in the specification.
func Connect(n net.Conn, h Headers) (c *Connection, e error) {
	if e := h.Validate(); e != nil {
		return nil, e
	}
	ch := h.Clone()
	c = &Connection{netconn: n,
		input:     make(chan MessageData),
		output:    make(chan wiredata),
		connected: false,
		session:   "",
		protocol:  SPL_10,
		subs:      make(map[string]chan MessageData)}
	c.MessageData = c.input

	// Check that the cilent wants a version we support
	if e = c.checkClientVersions(h); e != nil {
		return c, e
	}

	// OK, put a CONNECT on the wire
	c.wtr = bufio.NewWriter(n)        // Create the writer
	c.wsd = make(chan bool)           // Make the writer shutdown channel
	go c.writer()                     // Start it
	f := Frame{CONNECT, ch, NULLBUFF} // Create actual CONNECT frame
	r := make(chan error)             // Make the error channel fo a write
	c.output <- wiredata{f, r}        // Send the CONNECT frame
	e = <-r                           // Retrieve any error
	//
	if e != nil {
		c.wsd <- true // Shutdown the writer, we are done with errors
		return c, e
	}
	//
	e = c.connectHandler(ch)
	if e != nil {
		c.wsd <- true // Shutdown the writer, we are done with errors
		return c, e
	}
	// We are connected
	c.rsd = make(chan bool)
	go c.reader()
	//
	return c, e
}

// Connection handler, one time use during initial connect.
// Handle broker response, react to version incompatabilities, set up session, 
// and if necessary initialize heart beats.
func (c *Connection) connectHandler(h Headers) (e error) {
	c.rdr = bufio.NewReader(c.netconn)
	b, e := c.rdr.ReadBytes(0)
	if e != nil {
		return e
	}
	f, e := connectResponse(string(b))
	if e != nil {
		return e
	}
	//
	c.ConnectResponse = &Message{f.Command, f.Headers, f.Body}
	if c.ConnectResponse.Command == ERROR {
		return ECONERR
	}
	//
	e = c.checkConnectedVersions(h, c.ConnectResponse.Headers)
	if e != nil {
		return e
	}
	//
	if s, ok := c.ConnectResponse.Headers.Contains("session"); ok {
		c.session = s
	}

	if c.protocol >= SPL_11 {
		e = c.initializeHeartBeats(h)
		if e != nil {
			return e
		}
	}

	c.connected = true
	return nil
}

// Check client version, one time use during initial connect.
func (c *Connection) checkClientVersions(h Headers) (e error) {
	w := h.Value("accept-version")
	if w == "" { // Not present, client wants 1.0
		return nil
	}
	v := strings.SplitN(w, ",", -1) //
	for _, sv := range v {
		if hasValue(supported, sv) {
			return nil // At least one is supported
		}
	}
	return EBADVERCLI
}

// Check connected versions, one time use during initial connect.
func (c *Connection) checkConnectedVersions(ch, sh Headers) (e error) {
	chw := ch.Value("accept-version")
	shr := sh.Value("version")

	if chw == shr && supported.Supported(shr) {
		c.protocol = shr
		return nil
	}

	if chw == "" && shr == "" { // Straight up 1.0
		return nil // c.protocol defaults to SPL_10
	}

	cv := strings.SplitN(chw, ",", -1) // Client requested versions

	if chw != "" && shr != "" {
		if hasValue(cv, shr) {
			c.protocol = shr
			return nil
		} else {
			return EBADVERCLI
		}
	}

	if chw != "" && shr == "" { // Client asked for something, server is pure 1.0
		if hasValue(cv, SPL_10) {
			return nil // c.protocol defaults to SPL_10
		}
	}

	//
	if !supported.Supported(shr) {
		return EBADVERSVR // Client and server agree, but we do not support it
	}

	c.protocol = shr // Could be anything we support
	return nil
}
