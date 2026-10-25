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

type Tuple[T any] struct {
	Field T
}

type Tuple2[T1 any, T2 any] struct {
	Field1 T1
	Field2 T2
}

type Tuple3[T1 any, T2 any, T3 any] struct {
	Field1 T1
	Field2 T2
	Field3 T3
}

type Tuple4[T1 any, T2 any, T3 any, T4 any] struct {
	Field1 T1
	Field2 T2
	Field3 T3
	Field4 T4
}

type Tuple5[T1 any, T2 any, T3 any, T4 any, T5 any] struct {
	Field1 T1
	Field2 T2
	Field3 T3
	Field4 T4
	Field5 T5
}

type Tuple6[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any] struct {
	Field1 T1
	Field2 T2
	Field3 T3
	Field4 T4
	Field5 T5
	Field6 T6
}
