package main

import (
	"bytes"
	"testing"
)

func TestStore(t *testing.T) {
	opts :=  StoreOPT{
		PathTranformFunc: HashPathBuilder,
	}
	storage := NewStorage(opts)

	data := bytes.NewBuffer([]byte("Hello! How are you ?"))

	if err := storage.StoreFile("my picture", data); err != nil {
		t.Error(err)
	}
}