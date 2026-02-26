// Copyright (c) 2026 Reiner Pröls
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// SPDX-License-Identifier: MIT
//
// Author: Reiner Pröls

package omap

import (
	"maps"
	"slices"
)

type OMap[K comparable, V any] struct {
	m    map[K]V
	l    []K
	tmpl []V
}

func NewOMap[K comparable, V any](number int) OMap[K, V] {
	om := OMap[K, V]{
		m:    make(map[K]V, number),
		l:    make([]K, 0, number),
		tmpl: nil,
	}
	return om
}

func (om *OMap[K, V]) Add(key K, value V) {
	om.m[key] = value
	om.l = append(om.l, key)
	om.tmpl = nil
}

func (om *OMap[K, V]) Insert(index int, key K, value V) {
	om.m[key] = value
	om.l = slices.Insert(om.l, index, key)
	om.tmpl = nil
}

func (om *OMap[K, V]) RemoveByIndex(index int) {
	k := om.l[index]
	om.l = slices.Delete(om.l, index, index+1)
	if slices.Index(om.l, k) < 0 {
		delete(om.m, k)
		om.tmpl = nil
	}
}

func (om *OMap[K, V]) RemoveByKey(key K) {
	delete(om.m, key)
	l := make([]K, 0, len(om.l))
	for i := range len(om.l) {
		if om.l[i] != key {
			l = append(l, om.l[i])
		}
	}
	om.l = l
	om.tmpl = nil
}

func (om *OMap[K, V]) ReplaceByKey(key K, value V) {
	om.m[key] = value
	om.tmpl = nil
}

func (om *OMap[K, V]) ReplaceByIndex(index int, value V) {
	om.m[om.l[index]] = value
	om.tmpl = nil
}

func (om *OMap[K, V]) Clear() {
	clear(om.m)
	om.l = om.l[:0]
	om.tmpl = nil
}

func (om *OMap[K, V]) GetByKey(key K) (V, bool) {
	v, ok := om.m[key]
	return v, ok
}

func (om *OMap[K, V]) GetByIndex(index int) V {
	return om.m[om.l[index]]
}

func (om *OMap[K, V]) GetValues() []V {
	if om.tmpl != nil {
		return om.tmpl
	}
	om.tmpl = make([]V, 0, len(om.l))
	for i := range len(om.l) {
		om.tmpl = append(om.tmpl, om.m[om.l[i]])
	}
	return om.tmpl
}

func (om *OMap[K, V]) GetKeys() []K {
	return slices.Collect(maps.Keys(om.m))
}

func (om *OMap[K, V]) Sort(cmp func(a, b V) int) {
	slices.SortFunc(om.l, func(a, b K) int {
		return cmp(om.m[a], om.m[b])
	})
	om.tmpl = nil
}

func (om *OMap[K, V]) Len() int {
	return len(om.l)
}

/*
func (om *OMap[K, V]) Find(value V) []int {
	list := make([]int, 0, len(om.l))
	for i := range len(om.l) {
		if m.m[om.l[i]] == value {
			list = append(list, i)
		}
	}
	return list
}
*/

func (om *OMap[K, V]) IterKey() <-chan K {
	ch := make(chan K)
	go func() {
		for _, k := range om.l {
			ch <- k
		}
		close(ch)
	}()
	return ch
}

func (om *OMap[K, V]) IterValue() <-chan V {
	ch := make(chan V)
	go func() {
		for _, k := range om.l {
			ch <- om.m[k]
		}
		close(ch)
	}()
	return ch
}
