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

type Time struct{ v, min, max time.Time }

func LimitedTime(v, min, max time.Time) *Time {
	return &Time{v, min, max}
}

func MustParseTime(s string) *Time {
	var v time.Time
	if err := v.UnmarshalText([]byte(s)); err != nil {
		panic(err)
	}
	return NewTime(v)
}

func MustParseLimitedTime(sv, smin, smax string) *Time {
	var v, min, max time.Time
	if err := v.UnmarshalText([]byte(sv)); err != nil {
		panic(err)
	}
	if err := min.UnmarshalText([]byte(smin)); err != nil {
		panic(err)
	}
	if err := max.UnmarshalText([]byte(smax)); err != nil {
		panic(err)
	}
	return LimitedTime(v, min, max)
}

func NewTime(v time.Time) *Time {
	return &Time{v: v}
}

func (opt Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(opt.String())
}

func (opt Time) MarshalText() ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v.MarshalText()
}

func (opt Time) MarshalYAML() (interface{}, error) {
	text, err := opt.MarshalText()
	return string(text), err
}

func (opt *Time) Set(s string) error {
	return opt.UnmarshalText([]byte(s))
}

func (opt *Time) Store(v time.Time) error {
	if !opt.min.Equal(opt.max) {
		if v.Before(opt.min) {
			return fmt.Errorf("too soon")
		}
		if v.After(opt.max) {
			return fmt.Errorf("too late")
		}
	}
	mutex.Lock()
	defer mutex.Unlock()
	opt.v = v
	publish(uintptr(unsafe.Pointer(opt)))
	return nil
}

func (opt Time) String() string {
	text, err := opt.MarshalText()
	if err != nil {
		return err.Error()
	}
	return string(text)
}

func (opt *Time) UnmarshalJSON(text []byte) error {
	var s string
	if err := json.Unmarshal(text, &s); err != nil {
		return err
	}
	return opt.Set(s)
}

func (opt *Time) UnmarshalText(text []byte) error {
	var v time.Time
	if err := v.UnmarshalText(text); err != nil {
		return err
	}
	return opt.Store(v)
}

func (opt Time) Value() time.Time {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v
}
