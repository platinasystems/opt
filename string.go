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

type String[T ~string] struct {
	v   T
	aka []T
}

// Alias returns a string option that assures a new string matches either
// primary name or one of the aliasws before it's update.
func Alias[T ~string](name T, aka ...T) *String[T] {
	return &String[T]{name, append(aka, name)}
}

func NewString[T ~string](v T) *String[T] {
	return &String[T]{v: v}
}

func (opt String[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(opt.Value())
}

func (opt String[T]) MarshalText() ([]byte, error) {
	return []byte(opt.String()), nil
}

func (opt String[T]) MarshalYAML() (interface{}, error) {
	return opt.Value(), nil
}

func (opt *String[T]) Set(s string) error {
	return opt.Store(T(s))
}

func (opt *String[T]) Store(v T) error {
	if len(opt.aka) == 0 {
		return opt.store(v)
	}
	for _, s := range opt.aka {
		if s == v {
			return opt.store(v)
		}
	}
	return fmt.Errorf("%q invalid", v)
}

func (opt *String[T]) store(v T) error {
	mutex.Lock()
	defer mutex.Unlock()
	opt.v = v
	publish(uintptr(unsafe.Pointer(opt)))
	return nil
}

func (opt String[T]) String() string {
	return string(opt.Value())
}

func (opt *String[T]) UnmarshalJSON(text []byte) error {
	var v T
	if err := json.Unmarshal(text, &v); err != nil {
		return err
	}
	return opt.Store(v)
}

func (opt *String[T]) UnmarshalText(text []byte) error {
	return opt.Set(string(text))
}

func (opt String[T]) Value() T {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v
}

type Strings[T ~string] struct{ v []T }

func NewStrings[T ~string](v []T) *Strings[T] {
	return &Strings[T]{v}
}

func (opt Strings[T]) MarshalJSON() ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	ses := make([]string, len(opt.v))
	for i, v := range opt.v {
		ses[i] = string(v)
	}
	return json.Marshal(ses)
}

func (opt Strings[T]) MarshalText() ([]byte, error) {
	return []byte(opt.String()), nil
}

func (opt Strings[T]) MarshalYAML() (interface{}, error) {
	return opt.Value(), nil
}

func (opt *Strings[T]) Store(v []T) error {
	mutex.Lock()
	defer mutex.Unlock()
	opt.v = v
	publish(uintptr(unsafe.Pointer(opt)))
	return nil
}

func (opt Strings[T]) String() string {
	sb := new(strings.Builder)
	mutex.RLock()
	defer mutex.RUnlock()
	fmt.Fprint(sb, "[")
	for i, s := range opt.v {
		if i > 0 {
			fmt.Fprint(sb, " ")
		}
		if strings.Contains(string(s), " \t") {
			fmt.Fprintf(sb, "%q", s)
		} else {
			fmt.Fprint(sb, s)
		}
	}
	fmt.Fprint(sb, "]")
	return sb.String()
}

func (opt *Strings[T]) UnmarshalTOML(input interface{}) error {
	l, ok := input.([]interface{})
	if !ok {
		return fmt.Errorf("%T invalid", input)
	}
	v := make([]T, len(l))
	for i, iv := range l {
		if s, ok := iv.(string); ok {
			v[i] = T(s)
		} else {
			return fmt.Errorf("[%d]{%T} invalid", i, iv)
		}
	}
	return opt.Store(v)
}

func (opt *Strings[T]) UnmarshalYAML(unmarshal func(interface{}) error) error {
	v := make([]T, 0)
	if err := unmarshal(&v); err != nil {
		return err
	}
	return opt.Store(v)
}

func (opt Strings[T]) Value() []T {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v
}
