// Copyright © 2021-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package opt

import (
	"os"
	"strings"
)

// Associate environment variable names with Opt[T].Set()
type Env map[string]func(string) error

// Set options from os.Environ() or given list of KEY=VALUEs.
func (env Env) Set(args ...string) error {
	if len(args) == 0 {
		args = os.Environ()
	}
	for _, k := range args {
		if len(k) == 0 {
			continue
		}
		var v string
		if eq := strings.Index(k, "="); eq > 0 {
			v = k[eq+1:]
			k = k[:eq]
		}
		if set, ok := env[k]; ok {
			if err := set(v); err != nil {
				return err
			}
		}
	}
	return nil
}
