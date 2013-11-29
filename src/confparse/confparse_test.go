package confparse

import (
	"testing"
)

func TestNew(t *testing.T) {
	conf, err := New("test.conf")
	if err != nil {
		t.Fatal(err)
	}

	v1, err := conf.GetInt("", "key1")
	if v1 != 10 || err != nil {
		t.Fatal("E")
	}

	v2, err := conf.GetStr("", "key2")
	if v2 != "test" || err != nil {
		t.Fatal("E")
	}

	v3, err := conf.GetStr("", "key3")
	if v3 != "test with space" || err != nil {
		t.Fatal("E")
	}

	v4, err := conf.GetInt("section1", "sec_key1")
	if v4 != 60 || err != nil {
		t.Fatal("E")
	}

	v5, err := conf.GetStr("section2", "sec_key2")
	if v5 != "hello" || err != nil {
		t.Fatal("E")
	}
}
