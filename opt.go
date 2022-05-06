// Copyright © 2021-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package opt provides exclusive access and parsed input of generic options.
package opt

import (
	"encoding"
	"sync"
)

var (
	mutex sync.RWMutex
	subs  []chan<- uintptr
)

// Subscribe to option change notifications.
func Subscribe(ch chan<- uintptr) {
	mutex.Lock()
	defer mutex.Unlock()
	subs = append(subs, ch)
}

// Unsubscribe to option change notifications.
func Unsubscribe(ch chan<- uintptr) {
	mutex.Lock()
	defer mutex.Unlock()
	for i, sub := range subs {
		if ch == sub {
			copy(subs[i:], subs[i+1:])
			subs = subs[:len(subs)-1]
			break
		}
	}
}

func publish(ptr uintptr) {
	for _, sub := range subs {
		sub <- ptr
	}
}

func textunmarshaler(v any) func([]byte) error {
	return v.(encoding.TextUnmarshaler).UnmarshalText
}
