package main

import (
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	os.RemoveAll(".generatenew")
	os.Mkdir(".generatenew", 0755)
	wd, _ := os.Getwd()
	os.Chdir(".generatenew")
	t.Log("--- GENERATE NEW WEBSITE website1")
	generateNewWebsite("website1", "https://website1.test", true, true, true)

	compare(".generatenew", "testdata/.generatenew", t)

	os.Chdir(wd)
	os.RemoveAll(".generatenew")
}

