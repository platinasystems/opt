// Copyright © 2021-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package opt

import (
	"encoding/json"
	"fmt"
	"time"
	"unsafe"
)

type Duration struct{ v, min, max time.Duration }

func LimitedDuration(v, min, max time.Duration) *Duration {
	return &Duration{v, min, max}
}

func MustParseDuration(s string) *Duration {
	v, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return &Duration{v: v}
}

func NewDuration(v time.Duration) *Duration {
	return &Duration{v: v}
}

func (opt Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(opt.String())
}

func (opt Duration) MarshalText() ([]byte, error) {
	return []byte(opt.String()), nil
}

func (opt Duration) MarshalYAML() (interface{}, error) {
	return opt.String(), nil
}

func (opt *Duration) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	return opt.Store(v)
}

func (opt *Duration) Store(v time.Duration) error {
	if opt.min != opt.max {
		if v < opt.min {
			return fmt.Errorf("%v < min{%v}", v, opt.min)
		}
		if v > opt.max {
			return fmt.Errorf("%v > max{%v}", v, opt.max)
		}
	}
	mutex.Lock()
	defer mutex.Unlock()
	opt.v = v
	publish(uintptr(unsafe.Pointer(opt)))
	return nil
}

func (opt Duration) String() string {
	return opt.Value().String()
}

func (opt *Duration) UnmarshalJSON(text []byte) error {
	var s string
	if err := json.Unmarshal(text, &s); err != nil {
		return err
	}
	return opt.Set(s)
}

func (opt *Duration) UnmarshalText(text []byte) error {
	return opt.Set(string(text))
}

func (opt Duration) Value() time.Duration {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v
}
