package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type BasicCrypto struct {
}

func (b *BasicCrypto) copyStream(stream cipher.Stream, dst io.Writer, src io.Reader) (int, error) {
	const bufferSize = 32 * 1024
	buf := make([]byte, bufferSize)
	totalWritten := 0

	for {
		n, readErr := src.Read(buf)
		if (n > 0) {
			stream.XORKeyStream(buf[:n], buf[:n])

			// Write the processed data to the destination
			nn, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return totalWritten, writeErr
			}
			totalWritten += nn
		}

		// Handle the end of the file
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return totalWritten, readErr
		}
	}
	return totalWritten, nil
}

// Encrypt encrypts data from src to dst using AES in CTR mode.
// It generates a random IV, writes it to the beginning of dst, initializes the AES CTR stream with the key and IV,
// and then uses the stream to encrypt the remaining data.
func (b *BasicCrypto) Encrypt(encryptionKey []byte, dst io.Writer, src io.Reader) (int, error) {
	cipherBlock, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return 0, err
	}

	// Generate a random IV for the AES CTR mode
	iv := make([]byte, cipherBlock.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	// Write the IV to the beginning of dst, so it can be used for decryption
	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}

	stream := cipher.NewCTR(cipherBlock, iv)
	return b.copyStream(stream, dst, src)
}

// copyDecrypt decrypts data from src to dst using AES in CTR mode.
// It reads an IV from the beginning of src, initializes the AES CTR stream with the key and IV,
// and then uses the stream to decrypt the remaining data.
func (b *BasicCrypto) Decrypt(encryptionKey []byte, dst io.Writer, src io.Reader) (int, error) {
	cipherBlock, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return 0, err
	}

	// Read the IV from the beginning of src
	iv := make([]byte, cipherBlock.BlockSize())
	if _, err := io.ReadFull(src, iv); err != nil {
		return 0, err
	}

	stream := cipher.NewCTR(cipherBlock, iv)
	return b.copyStream(stream, dst, src)
}