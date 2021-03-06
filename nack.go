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

// Nack a STOMP 1.1+ message. Headers must contain a "message-id" key and
// a "subscription" key.  Disallowed for STOMP 1.0.
func (c *Connection) Nack(h Headers) (e error) {
	c.log(NACK, "start")
	if !c.connected {
		return ECONBAD
	}
	if c.protocol == SPL_10 {
		return EBADVERNAK
	}
	_, e = checkHeaders(h, c)
	if e != nil {
		return e
	}
	if _, ok := h.Contains("subscription"); !ok {
		return EREQSUBNAK
	}
	if _, ok := h.Contains("message-id"); !ok {
		return EREQMIDNAK
	}
	ch := h.Clone()
	e = c.transmitCommon(NACK, ch)
	c.log(NACK, "end")
	return e
}
