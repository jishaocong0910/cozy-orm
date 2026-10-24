// Copyright 2024-present jishaocong0910
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

package orm

import (
	"runtime/debug"
	"slices"
)

func checkMust(must bool, err error) error {
	if must && err != nil {
		panic(err)
	}
	return err
}

func deferStack() []byte {
	stack := debug.Stack()
	line := 0
	begin := 0
	for i, b := range stack {
		if b == '\n' {
			line++
		}
		if line == 9 {
			begin = i + 1
			break
		}
	}
	return stack[begin:]
}

type set[T comparable] map[T]struct{}

func (s set[T]) contain(e T) bool {
	_, ok := s[e]
	return ok
}

func (s set[T]) add(es ...T) {
	for _, e := range es {
		s[e] = struct{}{}
	}
}

func (s set[T]) concat(sources ...set[T]) set[T] {
	sources = slices.DeleteFunc(sources, func(s set[T]) bool {
		return len(s) == 0
	})
	if len(sources) == 0 {
		return s
	}

	sets := make([]set[T], 0, 1+len(sources))
	sets = append(sets, s)
	size := len(s)
	for _, source := range sources {
		sets = append(sets, source)
		size += len(source)
	}

	m := make(set[T], size)
	for _, s2 := range sets {
		for e := range s2 {
			m[e] = struct{}{}
		}
	}
	return m
}

func newSet[T comparable](es ...T) set[T] {
	m := make(set[T], len(es))
	if len(es) > 0 {
		for _, e := range es {
			m[e] = struct{}{}
		}
	}
	return m
}

func newSetWithCap[T comparable](cap int) set[T] {
	return make(set[T], cap)
}
