package common

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCheckUnusedConfig(t *testing.T) {
	md := &mapstructure.Metadata{
		Unused: make([]string, 0),
	}

	err := CheckUnusedConfig(md)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	md.Unused = []string{"foo", "bar"}
	err = CheckUnusedConfig(md)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestDecodeConfig(t *testing.T) {
	type Local struct {
		Foo string
		Bar string
	}

	raws := []interface{}{
		map[string]interface{}{
			"foo": "bar",
		},
		map[string]interface{}{
			"bar": "baz",
			"baz": "what",
		},
	}

	var result Local
	md, err := DecodeConfig(&result, raws...)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result.Foo != "bar" {
		t.Fatalf("invalid: %#v", result.Foo)
	}

	if result.Bar != "baz" {
		t.Fatalf("invalid: %#v", result.Bar)
	}

	if md == nil {
		t.Fatal("metadata should not be nil")
	}

	if !reflect.DeepEqual(md.Unused, []string{"baz"}) {
		t.Fatalf("unused: %#v", md.Unused)
	}
}

func TestDownloadableURL(t *testing.T) {
	// Invalid URL: has hex code in host
	_, err := DownloadableURL("http://what%20.com")
	if err == nil {
		t.Fatal("expected err")
	}

	// Invalid: unsupported scheme
	_, err = DownloadableURL("ftp://host.com/path")
	if err == nil {
		t.Fatal("expected err")
	}

	// Valid: http
	u, err := DownloadableURL("HTTP://packer.io/path")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if u != "http://packer.io/path" {
		t.Fatalf("bad: %s", u)
	}
}

func TestDownloadableURL_FilePaths(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("tempfile err: %s", err)
	}
	defer os.Remove(tf.Name())
	tf.Close()

	// Relative filepath. We run this test in a func so that
	// the defers run right away.
	func() {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd err: %s", err)
		}

		err = os.Chdir(filepath.Dir(tf.Name()))
		if err != nil {
			t.Fatalf("chdir err: %s", err)
		}
		defer os.Chdir(wd)

		filename := filepath.Base(tf.Name())
		u, err := DownloadableURL(filename)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if u != fmt.Sprintf("file:///%s", filename) {
			t.Fatalf("unexpected: %s", u)
		}
	}()

	// Test some cases with and without a schema prefix
	for _, prefix := range []string{"", "file://"} {
		// Nonexistent file
		_, err = DownloadableURL(prefix + "i/dont/exist")
		if err == nil {
			t.Fatal("expected err")
		}

		// Good file
		u, err := DownloadableURL(prefix + tf.Name())
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if u != fmt.Sprintf("file://%s", tf.Name()) {
			t.Fatalf("unexpected: %s", u)
		}
	}
}