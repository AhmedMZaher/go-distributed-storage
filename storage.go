package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// HashPathBuilder generates a FileIdentifier for a given file name.
// It computes the SHA-1 hash of the file name, encodes it to a hexadecimal string,
// and splits the hash into segments to create a directory path.
func HashPathBuilder(fileName string) FileIdentifier {
	hash := sha1.Sum([]byte(fileName))	
	hashStr := hex.EncodeToString(hash[:]) // hash[:] convert into byte slice

	segmentSize := 6
	numSegments := len(hashStr) / segmentSize
	pathSegments := make([]string, numSegments)

	for i := 0; i < numSegments; i++ {
		start, end := i * segmentSize, (i * segmentSize) + segmentSize
		pathSegments[i] = hashStr[start:end]
	}

	return FileIdentifier{
		PathName: strings.Join(pathSegments, "/"),
		FileName: hashStr,
	}
}

type FileIdentifier struct {
	PathName string
	FileName string
}

type StoreOPT struct {
	PathTranformFunc func(string) FileIdentifier
}

type Storage struct {
	Config StoreOPT
}

func NewStorage(storeOPT StoreOPT) *Storage {
	return &Storage{
		Config: storeOPT,
	}
}

func BuildFilePath(fileIdentifier FileIdentifier) string {
	return fmt.Sprintf("%s/%s", fileIdentifier.PathName, fileIdentifier.FileName)
}

// StoreFile reads from the input stream and writes the data to a file.
// The file path is determined by applying a transformation to the fileName.
// It creates necessary directories if they don't exist.
func (s *Storage) StoreFile(fileName string, inputStream io.Reader) error{
	fileIdentifier := s.Config.PathTranformFunc(fileName)
	if err := os.MkdirAll(fileIdentifier.PathName, os.ModePerm); err != nil {
		return err
	}

	pathAndFileName := BuildFilePath(fileIdentifier)

	destinationFile, err := os.Create(pathAndFileName)
	if err != nil {
		return err
	}

	n, err := io.Copy(destinationFile, inputStream)
	if err != nil {
		return err
	}

	log.Printf("%d bytes written to %s", n, pathAndFileName)

	return nil
}