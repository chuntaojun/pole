// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pole-group/pole/utils"
)

func TestKeyTree_Range(t *testing.T) {

	seek := func(tree *keyTree) {
		err := tree.Range(func(n *node) error {
			v := utils.IFLazy(n.parent != nil, func() interface{} {
				return n.parent.val
			}, func() interface{} {
				return ""
			})

			fmt.Printf("val : %v, parent : %v\n", n.val, v)
			return nil
		})

		if err != nil {
			t.Error(err)
		}
	}

	tree, _ := newTestKeyTree(1, 10)
	fmt.Printf("============================ Test-1 =========================\n")
	seek(tree)

	fmt.Printf("============================ Test-2 =========================\n")
	tree.Delete("2")
	seek(tree)
}

func TestKeyTree_FindNearbyLeft(t *testing.T) {
	tree := newTestKeyTreeByVal([]int{5,3,1,4,8,7,9})

	levels := tree.SeekLevel()
	for _, level := range levels {
		for _, no := range level {
			fmt.Printf("%s ", no.val)
		}
		fmt.Printf("\n")
	}

	n := tree.FindNearbyLeft("5")

	assert.Equalf(t, "3", n.val, "5 nearby left is 3!")

	n = tree.FindNearbyLeft("3")

	assert.Equalf(t, "1", n.val, "3 nearby left is 1!")

	n = tree.FindNearbyLeft("1")

	assert.Equalf(t, "1", n.val, "1 nearby left is 1!")
}

func newTestKeyTree(start, end int) (*keyTree, []string) {

	fmt.Printf("create new tree, start : %d, end : %d\n", start, end)

	tree := newKeyTree()
	orderV := make([]string, 0, 0)

	for i := start; i < end; i++ {
		tree.Insert(fmt.Sprintf("%d", i))
		orderV = append(orderV, fmt.Sprintf("%d", i))
	}

	return tree, orderV
}


func newTestKeyTreeByVal(val []int) *keyTree {
	fmt.Printf("create new tree by val %#v\n", val)

	tree := newKeyTree()
	orderV := make([]string, 0, 0)

	for _, v := range val {
		tree.Insert(fmt.Sprintf("%d", v))
		orderV = append(orderV, fmt.Sprintf("%d", v))
	}

	return tree
}
