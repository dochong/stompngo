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
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Initialize heart beats if necessary and possible.  Return an error (possibly
// nil) to mainline of initialization can not complete.  Establish heartbeat
// send and receive routines as necessary.
func (c *Connection) initializeHeartBeats(ch Headers) (e error) {
	// Client wants Heartbeats ?
	vc, ok := ch.Contains("heart-beat")
	if !ok || vc == "0,0" {
		return nil
	}
	// Server wants Heartbeats ?
	vs, ok := c.ConnectResponse.Headers.Contains("heart-beat")
	if !ok || vs == "0,0" {
		return nil
	}
	// Work area, may or may not become connection heartbeat data
	w := &heartbeat_data{cx: 0, cy: 0, sx: 0, sy: 0,
		hbs: true, hbr: true, // possible reset later
		sti: 0, rti: 0,
		ls: 0, lr: 0}

	// Client specified values
	cp := strings.Split(vc, ",")
	if len(cp) != 2 { // S/B caught by the server first
		return Error("invalid client heart-beat header: " + vc)
	}
	w.cx, e = strconv.ParseInt(cp[0], 10, 64)
	if e != nil {
		return Error("non-numeric cx heartbeat value: " + cp[0])
	}
	w.cy, e = strconv.ParseInt(cp[1], 10, 64)
	if e != nil {
		return Error("non-numeric cy heartbeat value: " + cp[1])
	}

	// Server specified values
	sp := strings.Split(vs, ",")
	if len(sp) != 2 {
		return Error("invalid server heart-beat header: " + vc)
	}
	w.sx, e = strconv.ParseInt(sp[0], 10, 64)
	if e != nil {
		return Error("non-numeric sx heartbeat value: " + sp[0])
	}
	w.sy, e = strconv.ParseInt(sp[1], 10, 64)
	if e != nil {
		return Error("non-numeric sy heartbeat value: " + sp[1])
	}

	// Check for sending needed
	if w.cx == 0 || w.sy == 0 {
		w.hbs = false // 
	}

	// Check for receiving needed
	if w.sx == 0 || w.cy == 0 {
		w.hbr = false // 
	}

	if !w.hbs && !w.hbr {
		return nil // none required
	}

	c.hbd = w                   // OK, we are doing some kind of heartbeating
	ct := time.Now().UnixNano() // Prime current time

	if w.hbs { // Finish sender parameters if required
		sm := max(w.cx, w.sy)   // ticker interval, ms
		w.sti = 1000000 * sm    // ticker interval, ns
		w.ssd = make(chan bool) // add shutdown channel
		w.ls = ct               // Best guess at start
		// fmt.Println("start send ticker")
		go c.sendTicker()
	}

	if w.hbr { // Finish receiver parameters if required
		rm := max(w.sx, w.cy)   // ticker interval, ms
		w.rti = 1000000 * rm    // ticker interval, ns
		w.rsd = make(chan bool) // add shutdown channel
		w.lr = ct               // Best guess at start
		// fmt.Println("start receive ticker")
		go c.receiveTicker()
	}
	return nil
}

// The heart beat send watch dog.
func (c *Connection) sendTicker() {
	ticker := time.NewTicker(time.Duration(c.hbd.sti))
	q := false
	for {
		select {
		case ct := <-ticker.C:
			ld := ct.UnixNano() - c.hbd.ls
			c.log("HeartBeat send TIC", ct, c.hbd.ls, ld)
			if ld > (c.hbd.sti - (c.hbd.sti / 5)) { // swag minus to be tolerant
				c.log("HeartBeat send data")
				// Send a heartbeat
				f := Frame{"\n", Headers{}, NULLBUFF} // Heartbeat frame
				r := make(chan error)
				c.output <- wiredata{f, r}
				e := <-r
				if e != nil {
					fmt.Printf("Heartbeat Send Failure: %v\n", e)
				}
			}
		case q = <-c.hbd.ssd:
			ticker.Stop()
			break
		}
		if q {
			break
		}
	}
	return
}

// The heart beat receive watch dog.
func (c *Connection) receiveTicker() {
	ticker := time.NewTicker(time.Duration(c.hbd.rti))
	q := false
	for {
		select {
		case ct := <-ticker.C:
			ld := ct.UnixNano() - c.hbd.lr
			c.log("HeartBeat receive TIC", ct, c.hbd.lr, ld)
			if ld > (c.hbd.rti + (c.hbd.rti / 5)) { // swag plus to be tolerant
				c.log("HeartBeat Receive Read is dirty")
				c.Hbrf = true // Flag possible dirty connection
			} else {
				c.Hbrf = false // Reset
			}
		case q = <-c.hbd.rsd:
			ticker.Stop()
			break
		}
		if q {
			break
		}
	}
	return
}
