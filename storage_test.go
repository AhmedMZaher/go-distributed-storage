package main

import (
	"bytes"
	"io"
	"testing"
)

func TestStore(t *testing.T) {
	opts :=  StoreOPT{
		PathTranformFunc: HashPathBuilder,
	}
	storage := NewStorage(opts)
	cleanup(t, storage)

	fileName := "my picture1"
	data := []byte("Hello! How are you ?")

	if err := storage.StoreFile(fileName, bytes.NewBuffer(data)); err != nil {
		t.Error(err)
	}

	r, err := storage.ReadFile(fileName)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)
	if string(b) != string(data) {
		t.Error("Wrong data Mismatch!")
	}
	
	if err := storage.DeleteFile(fileName); err != nil {
		t.Error(err)
	}
}

func cleanup(t *testing.T, s *Storage) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}