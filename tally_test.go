package main

import (
	"os"
	"reflect"
	"tally/cmd"
	"testing"
)

func TestCountSingleFile(t *testing.T) {
	res := make(map[string]int)
	expected := make(map[string]int)

	expected["baz"] = 1
	expected["foo"] = 4
	expected["bar"] = 2
	file, err := os.Open("test.txt")
	if err != nil {
		t.Fatalf("err %v", err)
	}
	defer file.Close()

	cmd.CountSingleFile(file, res)
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("does not match")
	}
}
