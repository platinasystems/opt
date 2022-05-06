// Copyright © 2021-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package opt

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

type StructExample struct {
	Scalar struct {
		Bool     Bool
		String   String[string]
		Int      Number[int]
		Float    Number[float64]
		Addr     Addr
		AddrPort AddrPort `toml:"addr_port" yaml:"addr_port"`
		Prefix   Prefix
		Duration Duration
		URL      URL
	}
	Slice struct {
		Ints     Numbers[int]
		Floats   Numbers[float64]
		Strings  Strings[string]
		Prefixes Prefixes
		Structs  []struct {
			Number Number[int]
			String String[string]
		}
	}
}

func ExampleMustParse() {
	fmt.Println(MustParseDuration("10s"))
	fmt.Println(MustParseAddr("192.168.0.1"))
	fmt.Println(MustParseAddrPort("192.168.0.1:80"))
	fmt.Println(MustParsePrefix("192.168.0.1/24"))
	fmt.Println(MustParseURL("https://golang.org/pkg/flag"))
	fmt.Println(MustParseTime("2006-01-02T15:04:05Z"))
	fmt.Println(MustParseTime("2006-01-02T15:04:05.999999999Z"))
	// Output:
	// 10s
	// 192.168.0.1
	// 192.168.0.1:80
	// 192.168.0.1/24
	// https://golang.org/pkg/flag
	// 2006-01-02T15:04:05Z
	// 2006-01-02T15:04:05.999999999Z
}

func ExampleAlias() {
	fmt.Print(Alias[string]("Thomas", "Tom", "Tommy").Set("Tommey"))
	// Output: "Tommey" invalid
}

func ExampleLimitedDuration() {
	fmt.Print(LimitedDuration(3*time.Second, 1*time.Second, 5*time.Second).
		Store(10 * time.Second))
	// Output: 10s > max{5s}
}

func ExampleLimitedNumber() {
	fmt.Print(LimitedNumber[uint](100, 100, 200).
		Store(201))
	// Output: 201 > max{200}
}

func ExampleLimitedTime() {
	fmt.Print(MustParseLimitedTime(
		"2006-01-02T15:04:05Z",
		"2006-01-02T12:00:00Z",
		"2006-01-02T17:00:00Z",
	).UnmarshalText([]byte(
		"2006-01-02T17:00:01Z")))
	// Output: too late
}

func ExampleEnv() {
	var x StructExample
	env := Env{
		"BOOL":      x.Scalar.Bool.Set,
		"STRING":    x.Scalar.String.Set,
		"INT":       x.Scalar.Int.Set,
		"FLOAT":     x.Scalar.Float.Set,
		"DURATION":  x.Scalar.Duration.Set,
		"ADDR":      x.Scalar.Addr.Set,
		"ADDR_PORT": x.Scalar.AddrPort.Set,
		"PREFIX":    x.Scalar.Prefix.Set,
		"URL":       x.Scalar.URL.Set,
	}
	err := env.Set(
		"BOOL",
		"STRING=hello world",
		"INT=321",
		"FLOAT=3.21",
		"ADDR=10.1.1.1",
		"ADDR_PORT=10.1.1.1:80",
		"PREFIX=10.1.1.1/24",
		"DURATION=5m",
		"URL=https://golang.org/pkg/flag",
	)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(x.Scalar.Bool)
		fmt.Println(x.Scalar.String)
		fmt.Println(x.Scalar.Int)
		fmt.Println(x.Scalar.Float)
		fmt.Println(x.Scalar.Duration)
		fmt.Println(x.Scalar.Addr)
		fmt.Println(x.Scalar.AddrPort)
		fmt.Println(x.Scalar.Prefix)
		fmt.Println(x.Scalar.URL)
	}
	// Output:
	// true
	// hello world
	// 321
	// 3.21
	// 5m0s
	// 10.1.1.1
	// 10.1.1.1:80
	// 10.1.1.1/24
	// https://golang.org/pkg/flag
}

func ExampleFlag() {
	var foobar struct {
		foo, bar Bool
	}
	fs := flag.NewFlagSet("example", flag.ContinueOnError)
	fs.Var(&foobar.foo, "foo", "fooish")
	fs.Var(&foobar.bar, "bar", "barish")
	fs.Var(MustParseAddr("192.168.1.1"), "a", "address")
	fs.Var(MustParseAddrPort("192.168.1.1:80"), "ap", "addr:port")
	fs.Var(MustParseDuration("3s"), "d", "duration")
	fs.Var(NewNumber[float64](1.23), "f", "float")
	fs.Var(NewNumber[int](123), "i", "int")
	fs.Var(MustParsePrefix("192.168.1.1/24"), "p", "prefix")
	fs.Var(NewString[string]("hello world"), "s", "string")
	fs.Var(NewBool(false), "t", "bool")
	fs.Var(MustParseURL("https://platinasystems.com"), "u", "url")
	err := fs.Parse([]string{
		"-foo",
		"-bar",
		"-a", "127.0.0.1",
		"-ap", "127.0.0.1:7",
		"-d", "3h",
		"-i", "987",
		"-f", "9.87",
		"-p", "127.0.0.1/32",
		"-s", "bonjour le monde",
		"-t",
		"-u", "https://golang.org/pkg/flag",
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fs.Visit(func(f *flag.Flag) {
			fmt.Printf("%s: %v\n", f.Usage, f.Value)
		})
	}
	// Output:
	// address: 127.0.0.1
	// addr:port: 127.0.0.1:7
	// barish: true
	// duration: 3h0m0s
	// float: 9.87
	// fooish: true
	// int: 987
	// prefix: 127.0.0.1/32
	// string: bonjour le monde
	// bool: true
	// url: https://golang.org/pkg/flag
}

func ExampleSubscribe() {
	var updates int
	var n Number[int]
	pn := uintptr(unsafe.Pointer(&n))
	ch := make(chan uintptr, 4)
	Subscribe(ch)
	defer Unsubscribe(ch)
	go func() {
		defer close(ch)
		for _, mod := range []int{3, 2, 1} {
			n.Store(mod)
		}
	}()
	for p := range ch {
		if p != pn {
			fmt.Println("mismatch")
		} else {
			updates += 1
		}
	}
	fmt.Print(updates)
	// Output: 3
}

func Example_struct_init() {
	text, err := json.MarshalIndent(StructExample{}, "", "  ")
	if err != nil {
		fmt.Println(err)
	} else {
		os.Stdout.Write(text)
	}
	// Output:
	// {
	//   "Scalar": {
	//     "Bool": false,
	//     "String": "",
	//     "Int": 0,
	//     "Float": 0,
	//     "Addr": "invalid IP",
	//     "AddrPort": "invalid AddrPort",
	//     "Prefix": "invalid Prefix",
	//     "Duration": "0s",
	//     "URL": ""
	//   },
	//   "Slice": {
	//     "Ints": [],
	//     "Floats": [],
	//     "Strings": [],
	//     "Prefixes": [],
	//     "Structs": null
	//   }
	// }
}

func Example_struct_toml() {
	var x StructExample
	_, err := toml.Decode(`
[[opt]]

[scalar]
bool = true
string = "hello world"
int = 789
float = 7.89
addr = "192.168.2.2"
addr_port = "192.168.2.2:80"
prefix = "192.168.2.2/24"
duration = "15s"
url = "https://golang.org/pkg/flag"

[slice]
ints = [ 3, 2, 1 ]
floats = [ 3.3, 2.2, 1.1 ]
strings = [
        "hello world",
        "bonjour le monde",
        "こんにちは世界",
]
prefixes = [
        "192.168.1.1/32",
        "192.168.1.2/32",
        "192.168.1.3/32",
    ]
[[slice.structs]]
number = 123
string = "hello world"
[[slice.structs]]
number = 456
string = "bonjour le monde"
[[slice.structs]]
number = 789
string = "こんにちは世界"
`, &x)
	if err != nil {
		fmt.Println(err)
		return
	}
	text, err := json.MarshalIndent(&x, "", "  ")
	if err != nil {
		fmt.Println(err)
	} else {
		os.Stdout.Write(text)
	}
	// Output:
	// {
	//   "Scalar": {
	//     "Bool": true,
	//     "String": "hello world",
	//     "Int": 789,
	//     "Float": 7.89,
	//     "Addr": "192.168.2.2",
	//     "AddrPort": "192.168.2.2:80",
	//     "Prefix": "192.168.2.2/24",
	//     "Duration": "15s",
	//     "URL": "https://golang.org/pkg/flag"
	//   },
	//   "Slice": {
	//     "Ints": [
	//       3,
	//       2,
	//       1
	//     ],
	//     "Floats": [
	//       3.3,
	//       2.2,
	//       1.1
	//     ],
	//     "Strings": [
	//       "hello world",
	//       "bonjour le monde",
	//       "こんにちは世界"
	//     ],
	//     "Prefixes": [
	//       "192.168.1.1/32",
	//       "192.168.1.2/32",
	//       "192.168.1.3/32"
	//     ],
	//     "Structs": [
	//       {
	//         "Number": 123,
	//         "String": "hello world"
	//       },
	//       {
	//         "Number": 456,
	//         "String": "bonjour le monde"
	//       },
	//       {
	//         "Number": 789,
	//         "String": "こんにちは世界"
	//       }
	//     ]
	//   }
	// }
}

func Example_struct_yaml() {
	var x StructExample
	err := yaml.Unmarshal([]byte(`
scalar:
  bool: true
  string: hello world
  int: 789
  float: 7.89
  addr: 192.168.2.2
  addr_port: 192.168.2.2:80
  prefix: 192.168.2.2/24
  duration: "55s"
  url: https://golang.org/pkg/flag

slice:
  ints:
    - 3
    - 2
    - 1
  floats:
    - 3.3
    - 2.2
    - 1.1
  strings:
    - hello world
    - bonjour le monde
    - こんにちは世界
  prefixes:
    - 192.168.1.1/32
    - 192.168.1.2/32
    - 192.168.1.3/32
  structs:
    - number: 123
      string: hello world
    - number: 456
      string: bonjour le monde
    - number: 789
      string: こんにちは世界
`), &x)
	if err != nil {
		fmt.Println(err)
		return
	}
	text, err := yaml.Marshal(&x)
	if err != nil {
		fmt.Println(err)
	} else {
		os.Stdout.Write(text)
	}
	// Output:
	// scalar:
	//   bool: true
	//   string: hello world
	//   int: 789
	//   float: 7.89
	//   addr: 192.168.2.2
	//   addr_port: 192.168.2.2:80
	//   prefix: 192.168.2.2/24
	//   duration: 55s
	//   url: https://golang.org/pkg/flag
	// slice:
	//   ints:
	//   - 3
	//   - 2
	//   - 1
	//   floats:
	//   - 3.3
	//   - 2.2
	//   - 1.1
	//   strings:
	//   - hello world
	//   - bonjour le monde
	//   - こんにちは世界
	//   prefixes:
	//   - 192.168.1.1/32
	//   - 192.168.1.2/32
	//   - 192.168.1.3/32
	//   structs:
	//   - number: 123
	//     string: hello world
	//   - number: 456
	//     string: bonjour le monde
	//   - number: 789
	//     string: こんにちは世界
}
