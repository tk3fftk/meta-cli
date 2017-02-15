package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/urfave/cli"
)

// VERSION gets set by the build script via the LDFLAGS
var VERSION string

var mkdirAll = os.MkdirAll
var stat = os.Stat
var writeFile = ioutil.WriteFile
var readFile = ioutil.ReadFile

const metaFile = "meta.json"

func getMeta(key string, metaSpace string) {
	// TODO must fix correct path later
	metaFilePath := metaSpace + "/test.json" // + metaFile
	metaJson, err := readFile(metaFilePath)
	if err != nil {
		panic(err)
	}
	var metaInterface map[string]interface{}
	err = json.Unmarshal(metaJson, &metaInterface)
	if err != nil {
		panic(err)
	}

	result := metaInterface[key]
	switch result.(type) {
	case map[string]interface{}, []interface{}:
		resultJson, _ := json.Marshal(result)
		fmt.Printf("%v", string(resultJson[:]))
	case nil:
		fmt.Print("null")
	default:
		fmt.Printf("%v", result)
	}
}

func setMeta(key string, value string, metaSpace string) {
	metaFilePath := metaSpace + "/" + metaFile
	var metaInterface map[string]interface{}

	_, err := stat(metaFilePath)
	// Not exist directory
	if err != nil {
		setupDir(metaSpace)
		// Initialize interface if first setting meta
		metaInterface = make(map[string]interface{})
	} else {
		metaJson, _ := readFile(metaFilePath)
		// Exist meta.json
		if len(metaJson) != 0 {
			err = json.Unmarshal(metaJson, &metaInterface)
			if err != nil {
				panic(err)
			}
		} else {
			// Exist meta.json but it is empty
			metaInterface = make(map[string]interface{})
		}
	}

	key, parsedValue := parseMetaValue(key, value)
	fmt.Printf("parsed %+v", parsedValue)
	// valueがarrayになるとき、既存のarrayを消してしまわないようにもってくる必要がある
	metaInterface[key] = parsedValue

	resultJson, err := json.Marshal(metaInterface)

	err = writeFile(metaFilePath, resultJson, 0666)
	if err != nil {
		panic(err)
	}
}

func parseMetaValue(key string, value string) (string, interface{}) {
	fmt.Println("key", key)
	fmt.Println("value", value)

	for p, c := range key {
		fmt.Println(p, string([]rune{c}))
		if string([]rune{c}) == "[" {
			nextChar := key[p+1]
			if nextChar == []byte("]")[0] {
				// Value is array
				var a [1]interface{}
				key = key[0:p] + key[p+2:]
				key, a[0] = parseMetaValue(key, value)
				return key, a
			} else {
				// Value is array with index
				fmt.Println("array with index")
				metaIndex, _ := strconv.Atoi(string(key[p+1]))
				key = key[0:p]
				fmt.Println("shorten key", key)
				fmt.Println("metaIndex", metaIndex)
				var a []interface{}
				a = make([]interface{}, metaIndex+1)
				key, a[metaIndex] = parseMetaValue(key, value)
				return key, a
			}
		} else if string([]rune{c}) == "." {
			// Value is object
			childKey := key[p+1:]
			key = key[0:p]
			fmt.Println("childKey", childKey)
			var obj map[string]interface{}
			obj = make(map[string]interface{})
			childKey, tmpValue := parseMetaValue(childKey, value)
			obj[childKey] = tmpValue
			return key, obj
		}
	}
	/*
		bracketIndex := strings.Index(key, "[")
		dotIndex := strings.Index(key, ".")
			if key[bracketIndex+1] == []byte("]")[0] && bracketIndex < dotIndex {
				// Value is array
				var a [1]interface{}
				key = key[0:bracketIndex]
				key, a[0] = parseMetaValue(key, value)
				return key, a
			} else if dotIndex > 0 {
				// Value is object
				childKey := key[dotIndex+1:]
				key = key[0:dotIndex]
				fmt.Println("childKey", childKey)
				var obj map[string]interface{}
				obj = make(map[string]interface{})
				_, obj[childKey] = parseMetaValue(childKey, value)
				return key, obj
			} else if bracketIndex >= 0 {
				// Value is array with number
				fmt.Println("array with num")
				metaIndex, _ := strconv.Atoi(string(key[bracketIndex+1]))
				key = key[0:bracketIndex]
				fmt.Println("shorten key", key)
				fmt.Println("metaIndex", metaIndex)
				var a []interface{}
				a = make([]interface{}, metaIndex+1)
				key, a[metaIndex] = parseMetaValue(key, value)
				return key, a
			}
	*/
	// Value is int
	i, err := strconv.Atoi(value)
	if err == nil {
		return key, i
	}
	// Value is float
	f, err := strconv.ParseFloat(value, 64)
	if err == nil {
		return key, f
	}
	// Value is bool
	b, err := strconv.ParseBool(value)
	if err == nil {
		return key, b
	}
	// Value is string
	return key, value
}

// setupDir makes directory and json file for meta
func setupDir(metaSpace string) {
	err := mkdirAll(metaSpace, 0777)
	if err != nil {
		panic(err)
	}
	err = writeFile(metaSpace+"/"+metaFile, []byte(""), 0666)
	if err != nil {
		panic(err)
	}
}

var cleanExit = func() {
	os.Exit(0)
}

// finalRecover makes one last attempt to recover from a panic.
// This should only happen if the previous recovery caused a panic.
func finalRecover() {
	if p := recover(); p != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Something terrible has happened. Please file a ticket with this info:")
		fmt.Fprintf(os.Stderr, "ERROR: %v\n%v\n", p, string(debug.Stack()))
	}
	cleanExit()
}

func main() {
	defer finalRecover()

	var metaSpace string

	app := cli.NewApp()
	app.Name = "meta-cli"
	app.Usage = "get or set metadata for Screwdriver build"
	app.UsageText = "meta command arguments [options]"
	app.Copyright = "(c) 2017 Yahoo Inc."

	if VERSION == "" {
		VERSION = "0.0.0"
	}
	app.Version = VERSION

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "meta-space",
			Usage:       "Location of meta temporarily",
			Value:       "/sd/meta",
			Destination: &metaSpace,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "get",
			Usage: "Get a metadata with key",
			Action: func(c *cli.Context) error {
				if len(c.Args()) == 0 {
					return cli.ShowAppHelp(c)
				}
				setupDir(metaSpace)
				getMeta(c.Args().First(), metaSpace)
				return nil
			},
			Flags: app.Flags,
		},
		{
			Name:  "set",
			Usage: "Set a metadata with key and value",
			Action: func(c *cli.Context) error {
				if len(c.Args()) <= 1 {
					return cli.ShowAppHelp(c)
				}
				setMeta(c.Args().Get(0), c.Args().Get(1), metaSpace)
				return nil
			},
			Flags: app.Flags,
		},
	}

	app.Run(os.Args)
}
