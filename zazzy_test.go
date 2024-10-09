package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenameExt(t *testing.T) {
	if s := renameExt("foo.md", ".md", ".html"); s != "foo.html" {
		t.Error(s)
	}
	if s := renameExt("foo.md", "", ".html"); s != "foo.html" {
		t.Error(s)
	}
	if s := renameExt("foo.txt", ".md", ".html"); s != "foo.txt" {
		t.Error(s)
	}
	if s := renameExt("foo", ".md", ".html"); s != "foo" {
		t.Error(s)
	}
	if s := renameExt("foo", "", ".html"); s != "foo.html" {
		t.Error(s)
	}
}

func TestRunPlugins(t *testing.T) {
	// external command
	if s, err := run(Vars{}, "echo", "hello"); err != nil || s != "hello\n" {
		t.Error(s, err)
	}
	// passing variables to plugins
	if s, err := run(Vars{"foo": "bar"}, "sh", "-c", "echo $ZS_FOO"); err != nil || s != "bar\n" {
		t.Error(s, err)
	}

	// custom plugin overriding external command
	os.Mkdir(ZSDIR, 0755)
	script := `#!/bin/sh
echo foo
`
	os.WriteFile(filepath.Join(ZSDIR, "echo"), []byte(script), 0755)
	if s, err := run(Vars{}, "./echo", "hello"); err != nil || s != "foo\n" {
		t.Error(s, err)
	}
	os.Remove(filepath.Join(ZSDIR, "echo"))
	os.Remove(ZSDIR)
}

func TestVars(t *testing.T) {
	tests := map[string]Vars{
		`
foo: bar
title: Hello, world!
---
Some content in markdown
`: {
			"foo":       "bar",
			"title":     "Hello, world!",
			"url":       "test.html",
			"file":      "test.md",
			"output":    filepath.Join(PUBDIR, "test.html"),
			"__content": "Some content in markdown\n",
		},
		`
url: "example.com/foo.html"
---
Hello
`: {
			"url":       "example.com/foo.html",
			"__content": "Hello\n",
		},
	}

	for script, vars := range tests {
		os.WriteFile("test.md", []byte(script), 0644)
		if v, s, err := getVars("test.md", Vars{"baz": "123"}); err != nil {
			t.Error(err)
		} else if s != vars["__content"] {
			t.Error(s, vars["__content"])
		} else {
			for key, value := range vars {
				if key != "__content" && v[key] != value {
					t.Error(key, v[key], value)
				}
			}
		}
	}
}

func TestRender(t *testing.T) {
	vars := map[string]string{"foo": "bar"}

	if s, _ := render("foo bar", vars, 1); s != "foo bar" {
		t.Error(s)
	}
	if s, _ := render("a {{printf short}} text", vars, 1); s != "a short text" {
		t.Error(s)
	}
	if s, _ := render("{{printf Hello}} x{{foo}}z", vars, 1); s != "Hello xbarz" {
		t.Error(s)
	}
	// Test error case
	if _, err := render("a {{greet text ", vars, 1); err == nil {
		t.Error("error expected")
	}
}

func TestRenderFavicon2(t *testing.T) {

	// 1st time = download and cache
	os.RemoveAll(PUBDIR)
	if s, _ := render("{{favicon https://laurent.lourenco.pro}}", nil, 1); s != "<img src=\"/img/favicons/laurent-lourenco-pro+180x180+.png\" alt=\"icon\" class=\"favicon\" role=\"img\">" {
		t.Error(s)
	} else {
		// 2nd time = render from cache
		if s, _ := render("{{favicon https://laurent.lourenco.pro/}}", nil, 1); s != "<img src=\"/img/favicons/laurent-lourenco-pro+180x180+.png\" alt=\"icon\" class=\"favicon\" role=\"img\">" {
			t.Error(s)
		} else {
			os.RemoveAll(PUBDIR)
		}
	}
}
