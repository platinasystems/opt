// Copyright © 2021-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package opt

import (
	"encoding/json"
	"fmt"
	"net/url"
	"unsafe"
)

type URL struct{ v url.URL }

func MustParseURL(s string) *URL {
	p, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return &URL{*p}
}

func (opt URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(opt.String())
}

func (opt URL) MarshalText() ([]byte, error) {
	return []byte(opt.String()), nil
}

func (opt URL) MarshalYAML() (interface{}, error) {
	return opt.String(), nil
}

func (opt *URL) Set(s string) error {
	p, err := url.Parse(s)
	if err != nil {
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	opt.v = *p
	publish(uintptr(unsafe.Pointer(opt)))
	return nil
}

func (opt URL) String() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v.String()
}

func (opt *URL) UnmarshalJSON(text []byte) error {
	var v any
	if err := json.Unmarshal(text, &v); err != nil {
		return err
	}
	if s, ok := v.(string); ok {
		return opt.Set(s)
	}
	return fmt.Errorf("%T invalid", v)
}

func (opt *URL) UnmarshalText(text []byte) error {
	return opt.Set(string(text))
}

func (opt URL) Value() url.URL {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v
}
