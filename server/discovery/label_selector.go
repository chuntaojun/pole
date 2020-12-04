// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"strings"

	"github.com/Conf-Group/pole/utils"
)

// {operator : AND, expressions : [
//		{key : label.{key}, operator : = or > or <, value : if start with label.{key} mean consumer and provider about this value must be equals},
//		{key : label.{key}, operator : = or > or <, value : if start with label.{key} mean consumer and provider about this value must be equals}
//]}

type OpType string

const (
	OpForAND  OpType = "AND"
	OpForOR   OpType = "OR"
	OpForEQ   OpType = "="
	OpForLT   OpType = "<"
	OpForGT   OpType = ">"
	OpForIN   OpType = "in"
	OpForNoIN OpType = "no_in"
)

type Selector struct {
	Operator    string   `json:"operator"`
	Expressions []Record `json:"expressions"`
}

type Record struct {
	operator OpType
	valueA   interface{}
	valueB   interface{}
}

func NewRecord(opType OpType, a, b interface{}) Record {
	return Record{
		operator: opType,
		valueA:   a,
		valueB:   b,
	}
}

func (r Record) GetOperator() OpType {
	return r.operator
}

func (r Record) GetValueA() interface{} {
	return r.valueA
}

func (r Record) GetValueB() interface{} {
	return r.valueB
}

var operateManager map[OpType]Operate

func init() {
	operateManager = make(map[OpType]Operate)
	operateManager[OpForAND] = new(OperateForAND)
}

func ExecOperate(record Record) bool {
	op := record.operator
	if operate, exist := operateManager[op]; exist {
		return operate.Exec(record.valueA, record.valueB)
	}
	return false
}

type Operate interface {
	Exec(a, b interface{}) bool
}

type OperateForAND struct {
}

func (ofa *OperateForAND) Exec(a, b interface{}) bool {
	return a.(bool) && b.(bool)
}

type OperateForOR struct {
}

func (ofo *OperateForOR) Exec(a, b interface{}) bool {
	return a.(bool) || b.(bool)
}

type OperateForEQ struct {
}

func (ofe *OperateForEQ) Exec(a, b interface{}) bool {
	switch a.(type) {
	case byte:
		return a.(byte) == b.(byte)
	case rune:
		return a.(rune) == b.(rune)
	case int:
		return a.(int) == b.(int)
	case float32:
		return a.(float32) == b.(float32)
	case float64:
		return a.(float64) == b.(float64)
	case string:
		return strings.Compare(a.(string), b.(string)) == 0
	case utils.Comparator:
		return a.(utils.Comparator).Compare(b) == 0
	default:
		return false
	}
}

type OperateForLT struct {
}

func (ofl *OperateForLT) Exec(a, b interface{}) bool {
	switch a.(type) {
	case byte:
		return a.(byte) < b.(byte)
	case rune:
		return a.(rune) < b.(rune)
	case int:
		return a.(int) < b.(int)
	case float32:
		return a.(float32) < b.(float32)
	case float64:
		return a.(float64) < b.(float64)
	case string:
		return strings.Compare(a.(string), b.(string)) < 0
	case utils.Comparator:
		return a.(utils.Comparator).Compare(b) < 0
	default:
		return false
	}
}

type OperateForGT struct {
}

func (ofl *OperateForGT) Exec(a, b interface{}) bool {
	switch a.(type) {
	case byte:
		return a.(byte) > b.(byte)
	case rune:
		return a.(rune) > b.(rune)
	case int:
		return a.(int) > b.(int)
	case float32:
		return a.(float32) > b.(float32)
	case float64:
		return a.(float64) > b.(float64)
	case string:
		return strings.Compare(a.(string), b.(string)) > 0
	case utils.Comparator:
		return a.(utils.Comparator).Compare(b) > 0
	default:
		return false
	}
}
