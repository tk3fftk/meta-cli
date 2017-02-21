package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

const testDir = "./_test"
const testFilePath = testDir + "/" + metaFile

func TestMain(m *testing.M) {
	// setup functions
	setupDir(testDir)
	readFile = func(filename string) ([]byte, error) { return ioutil.ReadFile("./test/test.json") }
	writeFile = func(filename string, data []byte, perm os.FileMode) error {
		return ioutil.WriteFile(testFilePath, data, 0666)
	}
	/*
		printf = func(format string, a ...interface{}) (n int, err error) {
			stdout := new(bytes.Buffer)
			fmt.Printf("%v", a)
			fmt.Printf("\n%v", format)
			fmt.Printf("\n%v", stdout)
			return fmt.Fprintf(stdout, format, a)
		}
	*/
	// run test
	retCode := m.Run()
	// teardown functions
	os.RemoveAll(testDir)
	os.Exit(retCode)
}

func TestSetupDir(t *testing.T) {
	os.RemoveAll(testDir)

	setupDir(testDir)
	_, err := os.Stat(testFilePath)
	if err != nil {
		t.Errorf("could not create %s in %s", metaFile, testDir)
	}
}

func TestGetMeta(t *testing.T) {
	stdout := new(bytes.Buffer)
	getMeta("str", testDir, stdout)
	expected := []byte("fuga")
	if bytes.Compare(expected, stdout.Bytes()) != 0 {
		t.Fatalf("not matched. expected '%v', actual '%v'", string(expected[:]), string(stdout.Bytes()[:]))
	}

	stdout = new(bytes.Buffer)
	getMeta("bool", testDir, stdout)
	expected = []byte("true")
	if bytes.Compare(expected, stdout.Bytes()) != 0 {
		t.Fatalf("not matched. expected '%v', actual '%v'", string(expected[:]), string(stdout.Bytes()[:]))
	}

	stdout = new(bytes.Buffer)
	getMeta("int", testDir, stdout)
	expected = []byte("1")
	if bytes.Compare(expected, stdout.Bytes()) != 0 {
		t.Fatalf("not matched. expected '%v', actual '%v'", string(expected[:]), string(stdout.Bytes()[:]))
	}

	stdout = new(bytes.Buffer)
	getMeta("float", testDir, stdout)
	expected = []byte("1.5")
	if bytes.Compare(expected, stdout.Bytes()) != 0 {
		t.Fatalf("not matched. expected '%v', actual '%v'", string(expected[:]), string(stdout.Bytes()[:]))
	}

	stdout = new(bytes.Buffer)
	getMeta("obj", testDir, stdout)
	expected = []byte("{\"ccc\":\"ddd\",\"momo\":{\"toke\":\"toke\"}}")
	if bytes.Compare(expected, stdout.Bytes()) != 0 {
		t.Fatalf("not matched. expected '%v', actual '%v'", string(expected[:]), string(stdout.Bytes()[:]))
	}

	stdout = new(bytes.Buffer)
	getMeta("ary", testDir, stdout)
	expected = []byte("[\"aaa\",\"bbb\"]")
	if bytes.Compare(expected, stdout.Bytes()) != 0 {
		t.Fatalf("not matched. expected '%v', actual '%v'", string(expected[:]), string(stdout.Bytes()[:]))
	}

	stdout = new(bytes.Buffer)
	getMeta("nu", testDir, stdout)
	expected = []byte("null")
	if bytes.Compare(expected, stdout.Bytes()) != 0 {
		t.Fatalf("not matched. expected '%v', actual '%v'", string(expected[:]), string(stdout.Bytes()[:]))
	}
}

func TestSetMeta(t *testing.T) {
}

func TestParseMetaValue(t *testing.T) {
}
