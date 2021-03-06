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

// Begin a STOMP transaction. Headers must contain a "transaction" header key
// with a value that is not an empty string.
func (c *Connection) Begin(h Headers) (e error) {
	c.log(BEGIN, "start")
	if !c.connected {
		return ECONBAD
	}
	_, e = checkHeaders(h, c)
	if e != nil {
		return e
	}
	if _, ok := h.Contains("transaction"); !ok {
		return EREQTIDBEG
	}
	if h.Value("transaction") == "" {
		return EREQTIDBEG
	}
	ch := h.Clone()
	e = c.transmitCommon(BEGIN, ch)
	c.log(BEGIN, "end")
	return e
}
