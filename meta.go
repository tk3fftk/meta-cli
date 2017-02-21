package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
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
var fprintf = fmt.Fprintf

const metaFile = "meta.json"

// Get meta from file based on key
func getMeta(key string, metaSpace string, output io.Writer) {
	metaFilePath := metaSpace + "/" + metaFile
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
		fprintf(output, "%v", string(resultJson[:]))
	case nil:
		fprintf(output, "null")
	default:
		fprintf(output, "%v", result)
	}
}

// Store meta to file with key and value
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

	key, parsedValue := parseMetaValue(key, value, metaInterface)
	fmt.Printf("parsed %+v", parsedValue)
	//fmt.Println("\n", reflect.TypeOf(metaInterface[key]).Kind())
	// valueがarrayのとき、既存のarrayを消してしまわないようにもってくる必要がある
	metaInterface[key] = parsedValue

	resultJson, err := json.Marshal(metaInterface)

	err = writeFile(metaFilePath, resultJson, 0666)
	if err != nil {
		panic(err)
	}
}

// Parse arguments of meta-cli to JSON
func parseMetaValue(key string, value string, previousMeta interface{}) (string, interface{}) {
	fmt.Println("key", key)
	fmt.Println("value", value)

	for p, c := range key {
		fmt.Println(p, string([]rune{c}))
		if string([]rune{c}) == "[" {
			nextChar := key[p+1]
			if nextChar == []byte("]")[0] {
				// Value is array
				var a [1]interface{}
				key = key[0:p] + key[p+2:] // remove bracket[] from key
				key, a[0] = parseMetaValue(key, value, previousMeta)
				return key, a
			} else {
				// Value is array with index
				fmt.Println("array with index")
				var i int
				for i = p + 1; ; i++ {
					_, err := strconv.Atoi(string(key[i]))
					fmt.Println("i, key[i]", i, string(key[i]))
					if err != nil {
						break
					}
				}
				fmt.Println("i", i)
				fmt.Println("keykeykey", key[p+1:i])
				metaIndex, _ := strconv.Atoi(key[p+1 : i])
				key = key[0:p] + key[i+1:] // remove bracket[num] from key
				fmt.Println("shorten key", key)
				fmt.Println("metaIndex", metaIndex)

				// interface to map
				v := reflect.ValueOf(previousMeta)
				fmt.Printf("previousMeta %+v \n", previousMeta)
				var previousMetaMap map[string]interface{}
				previousMetaMap = make(map[string]interface{})
				var prevKey string

				if v.Kind() == reflect.Map {
					for i, k := range v.MapKeys() {
						fmt.Println("aaaaaa", i, k.Interface(), v.MapIndex(k))
						prevKey, _ = k.Interface().(string)
						previousMetaMap[prevKey] = v.MapIndex(k)
					}
				} else if v.Kind() == reflect.Array {
					fmt.Println("arry")
				}

				fmt.Printf("previousMetaMap %+v \n", previousMetaMap)
				fmt.Println("prevKey, previousMetaMap", prevKey, previousMetaMap[prevKey], reflect.TypeOf(previousMetaMap[prevKey]))

				// この時点でpreviousMetaMapには、元のjsonデータと同じものがはいっているはず

				fmt.Println("key, prevKey", key, prevKey)

				// meta[key] is not exist, create array with nil
				if previousMetaMap[prevKey] == nil {
					var a []interface{}
					a = make([]interface{}, metaIndex+1)
					key, a[metaIndex] = parseMetaValue(key, value, previousMetaMap[prevKey])
					return key, a
				} else {
					var a []interface{}
					a = make([]interface{}, metaIndex+1)
					fmt.Println("key, previousMeta", key, previousMetaMap[prevKey])
					prev := reflect.ValueOf(previousMetaMap[prevKey])
					fmt.Println("prev", prev)
					if metaIndex+1 > prev.Len() {
						a = make([]interface{}, metaIndex+1)
					} else {
						a = make([]interface{}, prev.Len())
					}
					key, a[metaIndex] = parseMetaValue(key, value, previousMetaMap[prevKey])

					for i := 0; i < prev.Len(); i++ {
						if i != metaIndex {
							a[i] = prev.Index(i).Interface()
							fmt.Println("i, prev[i]", i, prev.Index(i).Interface())
						}
					}

					fmt.Println("key a[metaIndex]", key, a[metaIndex])
					fmt.Println("previousMeta", key, previousMetaMap[key])
					return key, a
				}
			}
		} else if string([]rune{c}) == "." {
			// Value is object
			childKey := key[p+1:]
			key = key[0:p]
			fmt.Println("childKey", childKey)
			var obj map[string]interface{}
			obj = make(map[string]interface{})
			childKey, tmpValue := parseMetaValue(childKey, value, previousMeta)
			obj[childKey] = tmpValue
			return key, obj
		}
	}
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
				getMeta(c.Args().First(), metaSpace, os.Stdout)
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
