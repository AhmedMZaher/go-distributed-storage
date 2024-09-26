package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestStore(t *testing.T) {
	opts := StoreOPT{
		PathTranformFunc: HashPathBuilder,
	}
	storage := NewStorage(opts)
	cleanup(t, storage)

	for i := 0; i < 50; i++ {
		fileName := fmt.Sprintf("foo_%d", i)
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

		if f := storage.HasKey(fileName); !f {
			t.Error("Expected file key to exist for:", fileName)
		}

		if err := storage.DeleteFile(fileName); err != nil {
			t.Error(err)
		}

		if f := storage.HasKey(fileName); f {
			t.Error("Expected file key to exist for:", fileName)
		}
	}
}

func cleanup(t *testing.T, s *Storage) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
