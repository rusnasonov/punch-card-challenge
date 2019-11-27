package main

import (
	"testing"
	"bytes"
)

func TestPucnhCardEncoding(t *testing.T){
	val := "123456"
	expected := [][]byte{
		[]byte("###*########"),
		[]byte("####*#######"),
		[]byte("#####*######"),
		[]byte("######*#####"),
		[]byte("#######*####"),
		[]byte("########*###"),
	}
	encoded, err := punchCardEncoder(val)
	if err != nil {
		t.Error(err)
	}
	for i, r := range encoded {
		if bytes.Compare(r, expected[i]) != 0 {
			t.Error("Exprected", expected[i], "Got", r)
		}
	}
}