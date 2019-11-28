// Copyright © 2019 Alibaba Co. Ltd.
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

package encode

import (
	"encoding/base64"
	"strings"
	"unicode"
)

// B64Encode implements base64 encode for string if necessary.
func B64Encode(s string) string {
	if strings.Contains(s, "\n") || !isASCII(s) {
		return "{base64}" + base64.StdEncoding.EncodeToString([]byte(s))
	}
	return s
}

// isASCII indicates string contains only ASCII.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
