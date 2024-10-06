package main

import "io"

type Cipher interface {
	Encrypt([]byte, io.Writer, io.Reader) (int, error)
	Decrypt([]byte, io.Writer, io.Reader) (int, error)
}