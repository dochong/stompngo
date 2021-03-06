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

// Send a STOMP MESSAGE.  Headers must contain a "destination" header.
// The message body (payload) is a string, which may be empty.
func (c *Connection) Send(h Headers, b string) (e error) {
	c.log(SEND, "start", h)
	if !c.connected {
		return ECONBAD
	}
	_, e = checkHeaders(h, c)
	if e != nil {
		return e
	}
	if _, ok := h.Contains("destination"); !ok {
		return EREQDSTSND
	}
	ch := h.Clone()
	f := Frame{SEND, ch, []uint8(b)}
	r := make(chan error)
	c.output <- wiredata{f, r}
	e = <-r
	c.log(SEND, "end", ch)
	return e // nil or not
}
