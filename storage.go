package main

import (
	"bytes"
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

func DefaultPathBuilder (fileName string) FileIdentifier {
	return FileIdentifier{
		PathName: fileName,
		FileName: fileName,
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
	if storeOPT.PathTranformFunc == nil {
		storeOPT.PathTranformFunc = DefaultPathBuilder
	}
	return &Storage{
		Config: storeOPT,
	}
}

func (fileIdentifier *FileIdentifier) BuildFilePath() string {
	return fmt.Sprintf("%s/%s", fileIdentifier.PathName, fileIdentifier.FileName)
}
// ReadFile reads the contents of a specified file and returns an io.Reader.
// It uses a buffer to temporarily store the file data in memory rather than writing 
// directly to a file. This approach avoids issues related to file locking and 
// access contention, as the file can be closed immediately after reading.
// Storing data in a buffer allows for efficient access without the need to 
// maintain file handles open longer than necessary, which reduces the risk 
// of file access conflicts in concurrent environments.
func (s *Storage) ReadFile(fileName string) (io.Reader, error){
	f, err := s.ReadIntoFile(fileName)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, f)
	return buf, err
}

func (s *Storage) ReadIntoFile(fileName string) (io.ReadCloser, error){
	fileIdentifier := s.Config.PathTranformFunc(fileName)
	fullPath := fileIdentifier.BuildFilePath()
	return os.Open(fullPath)
}

// StoreFile reads from the input stream and writes the data to a file.
// The file path is determined by applying a transformation to the fileName.
// It creates necessary directories if they don't exist.
func (s *Storage) StoreFile(fileName string, inputStream io.Reader) error{
	fileIdentifier := s.Config.PathTranformFunc(fileName)
	if err := os.MkdirAll(fileIdentifier.PathName, os.ModePerm); err != nil {
		return err
	}

	fullPath := fileIdentifier.BuildFilePath()

	destinationFile, err := os.Create(fullPath)
	if err != nil {
		return err
	}

	n, err := io.Copy(destinationFile, inputStream)
	if err != nil {
		return err
	}

	log.Printf("%d bytes written to %s", n, fullPath)

	return nil
}