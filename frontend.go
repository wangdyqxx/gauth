// Copyright 2020 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package casbin

import (
	"bytes"
	"encoding/json"
)

func CasbinJsGetPermissionForUser(e IEnforcer, user string) (string, error) {
	model := e.GetModel()
	m := map[string]interface{}{}
	m["m"] = model.ToText()
	policies := make([][]string, 0)
	amap, ok := model.GetKey("p")
	if !ok {
		return "", nil
	}
	for ptype := range amap {
		policy := model.GetPolicy("p", ptype)
		for i := range policy {
			policies = append(policies, append([]string{ptype}, policy[i]...))
		}
	}
	m["p"] = policies
	result := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(result)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(m)
	return result.String(), err
}
