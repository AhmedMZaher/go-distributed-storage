package main

import "io"

type cipher interface {
	encrypt([]byte, io.Writer, io.Reader) (int64, error)
	decrypt([]byte, io.Writer, io.Reader) (int64, error)
}