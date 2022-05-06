// Copyright © 2021-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package opt

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

type Bool struct{ v bool }

func NewBool(v bool) *Bool { return &Bool{v} }

func (opt Bool) IsBoolFlag() bool { return true }

func (opt Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(opt.Value())
}

func (opt Bool) MarshalText() ([]byte, error) {
	return []byte(opt.String()), nil
}

func (opt Bool) MarshalYAML() (interface{}, error) {
	return opt.Value(), nil
}

func (opt *Bool) Set(s string) error {
	var v bool
	if len(s) > 0 {
		if _, err := fmt.Sscan(s, &v); err != nil {
			return nil
		}
	} else {
		v = true
	}
	return opt.Store(v)
}

func (opt *Bool) Store(v bool) error {
	mutex.Lock()
	defer mutex.Unlock()
	opt.v = v
	publish(uintptr(unsafe.Pointer(opt)))
	return nil
}

func (opt Bool) String() string {
	return fmt.Sprint(opt.Value())
}

func (opt *Bool) UnmarshalJSON(text []byte) error {
	var v bool
	if err := json.Unmarshal(text, &v); err != nil {
		return err
	}
	return opt.Store(v)
}

func (opt *Bool) UnmarshalText(text []byte) error {
	return opt.Set(string(text))
}

func (opt Bool) Value() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v
}
