// Copyright © 2021-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package opt

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"unsafe"
)

type NetIP[T netip.Addr | netip.AddrPort | netip.Prefix] struct{ v T }

type Addr = NetIP[netip.Addr]
type AddrPort = NetIP[netip.AddrPort]
type Prefix = NetIP[netip.Prefix]

func MustParseAddr(s string) *NetIP[netip.Addr] {
	v, err := netip.ParseAddr(s)
	if err != nil {
		panic(err)
	}
	return &Addr{v}
}

func MustParseAddrPort(s string) *NetIP[netip.AddrPort] {
	v, err := netip.ParseAddrPort(s)
	if err != nil {
		panic(err)
	}
	return &AddrPort{v}
}

func MustParsePrefix(s string) *NetIP[netip.Prefix] {
	v, err := netip.ParsePrefix(s)
	if err != nil {
		panic(err)
	}
	return &Prefix{v}
}

func (opt NetIP[T]) Format(f fmt.State, verb rune) {
	format := string([]rune{'%', verb})
	fmt.Fprintf(f, format, opt.String())
}

func (opt NetIP[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(opt.String())
}

func (opt NetIP[T]) MarshalText() ([]byte, error) {
	return []byte(opt.String()), nil
}

func (opt NetIP[T]) MarshalYAML() (interface{}, error) {
	return opt.String(), nil
}

func (opt *NetIP[T]) Set(s string) error {
	return opt.UnmarshalText([]byte(s))
}

func (opt NetIP[T]) String() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return fmt.Sprint(opt.v)
}

func (opt *NetIP[T]) UnmarshalJSON(text []byte) error {
	var v any
	if err := json.Unmarshal(text, &v); err != nil {
		return err
	}
	if s, ok := v.(string); ok {
		return opt.Set(s)
	}
	return fmt.Errorf("%T invalid", v)
}

func (opt *NetIP[T]) UnmarshalText(text []byte) error {
	var v T
	if err := textunmarshaler(&v)(text); err != nil {
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	opt.v = v
	publish(uintptr(unsafe.Pointer(opt)))
	return nil
}

func (opt NetIP[T]) Value() T {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v
}

type NetIPs[T netip.Addr | netip.AddrPort | netip.Prefix] struct{ v []T }

type Addrs = NetIPs[netip.Addr]
type AddrPorts = NetIPs[netip.AddrPort]
type Prefixes = NetIPs[netip.Prefix]

func NewAddrs(v []netip.Addr) *NetIPs[netip.Addr] {
	return &NetIPs[netip.Addr]{v}
}

func NewAddrPorts(v []netip.AddrPort) *NetIPs[netip.AddrPort] {
	return &NetIPs[netip.AddrPort]{v}
}

func NewPrefixes(v []netip.Prefix) *NetIPs[netip.Prefix] {
	return &NetIPs[netip.Prefix]{v}
}

func (opt NetIPs[T]) MarshalJSON() ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	ses := make([]string, len(opt.v))
	for i, v := range opt.v {
		ses[i] = fmt.Sprint(v)
	}
	return json.Marshal(ses)
}

func (opt NetIPs[T]) MarshalYAML() (interface{}, error) {
	return opt.Value(), nil
}

func (opt *NetIPs[T]) Store(v []T) error {
	mutex.Lock()
	defer mutex.Unlock()
	opt.v = v
	publish(uintptr(unsafe.Pointer(opt)))
	return nil
}

func (opt *NetIPs[T]) UnmarshalTOML(input interface{}) error {
	l, ok := input.([]interface{})
	if !ok {
		return fmt.Errorf("%T invalid", input)
	}
	vs := make([]T, len(l))
	for i, iv := range l {
		s, ok := iv.(string)
		if !ok {
			return fmt.Errorf("[%d]{%T} invalid", i, iv)
		}
		if err := textunmarshaler(&vs[i])([]byte(s)); err != nil {
			return err
		}
	}
	return opt.Store(vs)
}

func (opt *NetIPs[T]) UnmarshalYAML(unmarshal func(interface{}) error) error {
	ses := make([]string, 0)
	if err := unmarshal(&ses); err != nil {
		return err
	}
	vs := make([]T, len(ses))
	for i, s := range ses {
		if err := textunmarshaler(&vs[i])([]byte(s)); err != nil {
			return err
		}
	}
	return opt.Store(vs)
}

func (opt NetIPs[T]) Value() []T {
	mutex.RLock()
	defer mutex.RUnlock()
	return opt.v
}
