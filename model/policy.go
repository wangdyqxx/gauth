// Copyright 2017 The casbin Authors. All Rights Reserved.
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

package model

import (
	"strconv"
	"strings"

	"github.com/wangdyqxx/casbinplus/rbac"
	"github.com/wangdyqxx/casbinplus/util"
)

type (
	PolicyOp int
)

const (
	PolicyAdd PolicyOp = iota
	PolicyRemove
)

const DefaultSep = ","

// BuildIncrementalRoleLinks provides incremental build the role inheritance relations.
func (model *Model) BuildIncrementalRoleLinks(rmMap map[string]rbac.RoleManager, op PolicyOp, sec string, ptype string, rules [][]string) error {
	if sec == "g" {
		amap, ok := model.GetKey(sec)
		if !ok {
			return nil
		}
		return amap[ptype].buildIncrementalRoleLinks(rmMap[ptype], op, rules)
	}
	return nil
}

// BuildRoleLinks initializes the roles in RBAC.
func (model *Model) BuildRoleLinks(rmMap map[string]rbac.RoleManager) error {
	model.PrintPolicy()
	amap, ok := model.GetKey("g")
	if !ok {
		return nil
	}
	for ptype, ast := range amap {
		rm := rmMap[ptype]
		err := ast.buildRoleLinks(rm)
		if err != nil {
			return err
		}
	}

	return nil
}

// PrintPolicy prints the policy to log.
func (model *Model) PrintPolicy() {
	if !model.GetLogger().IsEnabled() {
		return
	}

	policy := make(map[string][][]string)
	amap, ok := model.GetKey("p")
	if !ok {
		return
	}
	for key, ast := range amap {
		value, found := policy[key]
		if found {
			value = append(value, ast.Policy...)
			policy[key] = value
		} else {
			policy[key] = ast.Policy
		}
	}
	amap, ok = model.GetKey("g")
	if !ok {
		return
	}
	for key, ast := range amap {
		value, found := policy[key]
		if found {
			value = append(value, ast.Policy...)
			policy[key] = value
		} else {
			policy[key] = ast.Policy
		}
	}

	model.GetLogger().LogPolicy(policy)
}

// ClearPolicy clears all current policy.
func (model *Model) ClearPolicy() {
	amap, ok := model.GetKey("p")
	if !ok {
		return
	}
	for _, ast := range amap {
		ast.Policy = nil
		ast.PolicyMap = map[string]int{}
	}
	amap, ok = model.GetKey("g")
	if !ok {
		return
	}
	for _, ast := range amap {
		ast.Policy = nil
		ast.PolicyMap = map[string]int{}
	}
}

// GetPolicy gets all rules in a policy.
func (model *Model) GetPolicy(sec string, ptype string) [][]string {
	amap, ok := model.GetKey(sec)
	if !ok {
		return nil
	}
	return amap[ptype].Policy
}

// GetFilteredPolicy gets rules based on field filters from a policy.
func (model *Model) GetFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) [][]string {
	res := [][]string{}
	amap, ok := model.GetKey(sec)
	if !ok {
		return res
	}
	for _, rule := range amap[ptype].Policy {
		matched := true
		for i, fieldValue := range fieldValues {
			if fieldValue != "" && rule[fieldIndex+i] != fieldValue {
				matched = false
				break
			}
		}

		if matched {
			res = append(res, rule)
		}
	}

	return res
}

// HasPolicy determines whether a model has the specified policy rule.
func (model *Model) HasPolicy(sec string, ptype string, rule []string) bool {
	amap, ok := model.GetKey(sec)
	if !ok {
		return ok
	}
	_, ok = amap[ptype].PolicyMap[strings.Join(rule, DefaultSep)]
	return ok
}

// HasPolicies determines whether a model has any of the specified policies. If one is found we return true.
func (model *Model) HasPolicies(sec string, ptype string, rules [][]string) bool {
	for i := 0; i < len(rules); i++ {
		if model.HasPolicy(sec, ptype, rules[i]) {
			return true
		}
	}

	return false
}

// AddPolicy adds a policy rule to the model.
func (model *Model) AddPolicy(sec string, ptype string, rule []string) {

	amap, ok := model.GetKey(sec)
	if !ok {
		return
	}

	assertion := amap[ptype]
	assertion.Policy = append(assertion.Policy, rule)
	assertion.PolicyMap[strings.Join(rule, DefaultSep)] = len(amap[ptype].Policy) - 1

	if sec == "p" && assertion.priorityIndex >= 0 {
		if idxInsert, err := strconv.Atoi(rule[assertion.priorityIndex]); err == nil {
			i := len(assertion.Policy) - 1
			for ; i > 0; i-- {
				idx, err := strconv.Atoi(assertion.Policy[i-1][assertion.priorityIndex])
				if err != nil {
					break
				}
				if idx > idxInsert {
					assertion.Policy[i] = assertion.Policy[i-1]
					assertion.PolicyMap[strings.Join(assertion.Policy[i-1], DefaultSep)]++
				} else {
					break
				}
			}
			assertion.Policy[i] = rule
			assertion.PolicyMap[strings.Join(rule, DefaultSep)] = i
		}
	}
}

// AddPolicies adds policy rules to the model.
func (model *Model) AddPolicies(sec string, ptype string, rules [][]string) {
	_ = model.AddPoliciesWithAffected(sec, ptype, rules)
}

// AddPoliciesWithAffected adds policy rules to the model, and returns effected rules.
func (model *Model) AddPoliciesWithAffected(sec string, ptype string, rules [][]string) [][]string {
	var effected [][]string
	amap, ok := model.GetKey(sec)
	if !ok {
		return effected
	}
	for _, rule := range rules {
		hashKey := strings.Join(rule, DefaultSep)
		_, ok = amap[ptype].PolicyMap[hashKey]
		if ok {
			continue
		}
		effected = append(effected, rule)
		model.AddPolicy(sec, ptype, rule)
	}
	return effected
}

// RemovePolicy removes a policy rule from the model.
// Deprecated: Using AddPoliciesWithAffected instead.
func (model *Model) RemovePolicy(sec string, ptype string, rule []string) bool {
	amap, ok := model.GetKey(sec)
	if !ok {
		return ok
	}
	index, ok := amap[ptype].PolicyMap[strings.Join(rule, DefaultSep)]
	if !ok {
		return false
	}

	amap[ptype].Policy = append(amap[ptype].Policy[:index], amap[ptype].Policy[index+1:]...)
	delete(amap[ptype].PolicyMap, strings.Join(rule, DefaultSep))
	for i := index; i < len(amap[ptype].Policy); i++ {
		amap[ptype].PolicyMap[strings.Join(amap[ptype].Policy[i], DefaultSep)] = i
	}

	return true
}

// UpdatePolicy updates a policy rule from the model.
func (model *Model) UpdatePolicy(sec string, ptype string, oldRule []string, newRule []string) bool {
	oldPolicy := strings.Join(oldRule, DefaultSep)
	amap, ok := model.GetKey(sec)
	if !ok {
		return ok
	}
	index, ok := amap[ptype].PolicyMap[oldPolicy]
	if !ok {
		return false
	}

	amap[ptype].Policy[index] = newRule
	delete(amap[ptype].PolicyMap, oldPolicy)
	amap[ptype].PolicyMap[strings.Join(newRule, DefaultSep)] = index

	return true
}

// UpdatePolicies updates a policy rule from the model.
func (model *Model) UpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) bool {
	rollbackFlag := false
	// index -> []{oldIndex, newIndex}
	modifiedRuleIndex := make(map[int][]int)
	amap, ok := model.GetKey(sec)
	if !ok {
		return ok
	}
	// rollback
	defer func() {
		if rollbackFlag {
			for index, oldNewIndex := range modifiedRuleIndex {
				amap[ptype].Policy[index] = oldRules[oldNewIndex[0]]
				oldPolicy := strings.Join(oldRules[oldNewIndex[0]], DefaultSep)
				newPolicy := strings.Join(newRules[oldNewIndex[1]], DefaultSep)
				delete(amap[ptype].PolicyMap, newPolicy)
				amap[ptype].PolicyMap[oldPolicy] = index
			}
		}
	}()

	newIndex := 0
	for oldIndex, oldRule := range oldRules {
		oldPolicy := strings.Join(oldRule, DefaultSep)
		index, ok := amap[ptype].PolicyMap[oldPolicy]
		if !ok {
			rollbackFlag = true
			return false
		}

		amap[ptype].Policy[index] = newRules[newIndex]
		delete(amap[ptype].PolicyMap, oldPolicy)
		amap[ptype].PolicyMap[strings.Join(newRules[newIndex], DefaultSep)] = index
		modifiedRuleIndex[index] = []int{oldIndex, newIndex}
		newIndex++
	}

	return true
}

// RemovePolicies removes policy rules from the model.
func (model *Model) RemovePolicies(sec string, ptype string, rules [][]string) bool {
	effected := model.RemovePoliciesWithEffected(sec, ptype, rules)
	return len(effected) != 0
}

// RemovePoliciesWithEffected removes policy rules from the model, and returns effected rules.
func (model *Model) RemovePoliciesWithEffected(sec string, ptype string, rules [][]string) [][]string {
	var effected [][]string
	amap, ok := model.GetKey(sec)
	if !ok {
		return effected
	}
	for _, rule := range rules {
		index, ok := amap[ptype].PolicyMap[strings.Join(rule, DefaultSep)]
		if !ok {
			continue
		}

		effected = append(effected, rule)
		amap[ptype].Policy = append(amap[ptype].Policy[:index], amap[ptype].Policy[index+1:]...)
		delete(amap[ptype].PolicyMap, strings.Join(rule, DefaultSep))
		for i := index; i < len(amap[ptype].Policy); i++ {
			amap[ptype].PolicyMap[strings.Join(amap[ptype].Policy[i], DefaultSep)] = i
		}
	}
	return effected
}

// RemoveFilteredPolicy removes policy rules based on field filters from the model.
func (model *Model) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, [][]string) {
	var tmp [][]string
	var effects [][]string
	res := false
	amap, ok := model.GetKey(sec)
	if !ok {
		return ok, effects
	}
	amap[ptype].PolicyMap = map[string]int{}

	for _, rule := range amap[ptype].Policy {
		matched := true
		for i, fieldValue := range fieldValues {
			if fieldValue != "" && rule[fieldIndex+i] != fieldValue {
				matched = false
				break
			}
		}

		if matched {
			effects = append(effects, rule)
		} else {
			tmp = append(tmp, rule)
			amap[ptype].PolicyMap[strings.Join(rule, DefaultSep)] = len(tmp) - 1
		}
	}

	if len(tmp) != len(amap[ptype].Policy) {
		amap[ptype].Policy = tmp
		res = true
	}

	return res, effects
}

// GetValuesForFieldInPolicy gets all values for a field for all rules in a policy, duplicated values are removed.
func (model *Model) GetValuesForFieldInPolicy(sec string, ptype string, fieldIndex int) []string {
	values := []string{}
	amap, ok := model.GetKey(sec)
	if !ok {
		return values
	}
	for _, rule := range amap[ptype].Policy {
		values = append(values, rule[fieldIndex])
	}

	util.ArrayRemoveDuplicates(&values)

	return values
}

// GetValuesForFieldInPolicyAllTypes gets all values for a field for all rules in a policy of all ptypes, duplicated values are removed.
func (model *Model) GetValuesForFieldInPolicyAllTypes(sec string, fieldIndex int) []string {
	values := []string{}
	amap, ok := model.GetKey(sec)
	if !ok {
		return values
	}
	for ptype := range amap {
		values = append(values, model.GetValuesForFieldInPolicy(sec, ptype, fieldIndex)...)
	}

	util.ArrayRemoveDuplicates(&values)

	return values
}
