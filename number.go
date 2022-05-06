// Copyright © 2021-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package opt

import (
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"
)

type Numeric interface {
	~float32 | ~float64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Number[T Numeric] struct{ v, min, max T }

func LimitedNumber[T Numeric](v, min, max T) *Number[T] {
	return &Number[T]{v, min, max}
}

func NewNumber[T Numeric](v T) *Number[T] {
	return &Number[T]{v: v}
}

func (opt Number[T]) MarshalJSON() ([]byte, error) {
	v := opt.Value()
	return json.Marshal(float64(v))
}

func (opt Number[T]) MarshalText() ([]byte, error) {
	return []byte(opt.String()), nil
}

func (opt Number[T]) MarshalYAML() (interface{}, error) {
	return opt.Value(), nil
}

func (opt *Number[T]) Set(s string) error {
	var v T
	_, err := fmt.Sscan(s, &v)
	if err != nil {
		return err
	}
	return opt.Store(v)
}

func (opt *Number[T]) Store(v T) error {
	if opt.min != opt.max {
		if v < opt.min {
			return fmt.Errorf("%v < min{%v}", v, opt.min)
		}
		if v > opt.min {
			return fmt.Errorf("%v > max{%v}", v, opt.max)
		}
	}
	mutex.Lock()
	defer mutex.Unlock()
	opt.v = v
	publish(uintptr(unsafe.Pointer(opt)))
	return nil
}

func (opt Number[T]) String() string {
	return fmt.Sprint(opt.Value())
}

func (opt *Number[T]) UnmarshalJSON(text []byte) error {
	var v float64
	if err := json.Unmarshal(text, &v); err != nil {
		return err
	}
	return opt.Store(T(v))
}

func (opt *Number[T]) UnmarshalText(text []byte) error {
	return opt.Set(string(text))
}

func (opt Number[T]) Value() T {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v
}

type Numbers[T Numeric] struct{ v []T }

func NewNumbers[T Numeric](v []T) *Numbers[T] {
	return &Numbers[T]{v}
}

func (opt Numbers[T]) MarshalJSON() ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	f := make([]float64, len(opt.v))
	for i, n := range opt.v {
		f[i] = float64(n)
	}
	return json.Marshal(f)
}

func (opt Numbers[T]) MarshalText() ([]byte, error) {
	return []byte(opt.String()), nil
}

func (opt Numbers[T]) MarshalYAML() (interface{}, error) {
	return opt.Value(), nil
}

func (opt *Numbers[T]) Store(v []T) error {
	mutex.Lock()
	defer mutex.Unlock()
	opt.v = v
	publish(uintptr(unsafe.Pointer(opt)))
	return nil
}

func (opt Numbers[T]) String() string {
	sb := new(strings.Builder)
	mutex.RLock()
	defer mutex.RUnlock()
	fmt.Fprint(sb, "[")
	for i, n := range opt.v {
		if i > 0 {
			fmt.Fprint(sb, ", ")
		}
		fmt.Fprint(sb, n)
	}
	fmt.Fprint(sb, "]")
	return sb.String()
}

func (opt *Numbers[T]) UnmarshalTOML(input interface{}) error {
	v := make([]T, 0)
	l, ok := input.([]interface{})
	if !ok {
		return fmt.Errorf("%T invalid", input)
	}
	for i, iv := range l {
		switch t := iv.(type) {
		case int64:
			v = append(v, T(t))
		case float64:
			v = append(v, T(t))
		default:
			return fmt.Errorf("[%d]{%T} invalid", i, t)
		}
	}
	return opt.Store(v)
}

func (opt *Numbers[T]) UnmarshalYAML(unmarshal func(interface{}) error) error {
	v := make([]T, 0)
	if err := unmarshal(&v); err != nil {
		return err
	}
	return opt.Store(v)
}

func (opt Numbers[T]) Value() []T {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v
}
