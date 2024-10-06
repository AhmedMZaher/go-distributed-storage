package main

import "io"

type Cipher interface {
	Encrypt([]byte, io.Writer, io.Reader) (int64, error)
	Decrypt([]byte, io.Writer, io.Reader) (int64, error)
}