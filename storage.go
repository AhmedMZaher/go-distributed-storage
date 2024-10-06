package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// DefaultRootFolderName represents the default name for the root folder
// where data will be stored in the distributed storage system.
const DefaultRootFolderName = "data"

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
		start, end := i*segmentSize, (i*segmentSize)+segmentSize
		pathSegments[i] = hashStr[start:end]
	}

	return FileIdentifier{
		PathName: strings.Join(pathSegments, "/"),
		FileName: hashStr,
	}
}

func DefaultPathBuilder(fileName string) FileIdentifier {
	return FileIdentifier{
		PathName: fileName,
		FileName: fileName,
	}
}

type FileIdentifier struct {
	PathName string
	FileName string
}

func (FileIdentifier *FileIdentifier) firstPathSegment() string {
	pathSegments := strings.Split(FileIdentifier.PathName, "/")
	if len(pathSegments) == 0 {
		return ""
	}
	return pathSegments[0]
}

func (fileIdentifier *FileIdentifier) BuildFilePath() string {
	return fmt.Sprintf("%s/%s", fileIdentifier.PathName, fileIdentifier.FileName)
}

type PathTranformSignature func(string) FileIdentifier
type StoreOPT struct {
	PathTranformFunc PathTranformSignature
	RootDir          string
}

type Storage struct {
	Config StoreOPT
}

func NewStorage(storeOPT StoreOPT) *Storage {
	if storeOPT.PathTranformFunc == nil {
		storeOPT.PathTranformFunc = DefaultPathBuilder
	}
	if storeOPT.RootDir == "" {
		storeOPT.RootDir = DefaultRootFolderName
	}
	storeOPT.RootDir = strings.ReplaceAll(storeOPT.RootDir, ":", "_")
	return &Storage{
		Config: storeOPT,
	}
}

func (s *Storage) Clear() error {
	return os.RemoveAll(s.Config.RootDir)
}

func (s *Storage) prependTheRoot(path string) string {
	return fmt.Sprintf("%s/%s", s.Config.RootDir, path)
}

// HasKey checks if a file with the given name exists in the storage.
func (s *Storage) HasKey(fileName string) bool {
	fileIdentifier := s.Config.PathTranformFunc(fileName)
	fullPathWithRoot := s.prependTheRoot(fileIdentifier.BuildFilePath())

	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Storage) DeleteFile(fileName string) error {
	fileIdentifier := s.Config.PathTranformFunc(fileName)

	defer func() {
		log.Printf("Deleted file or directory: %s", fileIdentifier.firstPathSegment())
	}()

	firstPathSegmentWithRoot := s.prependTheRoot(fileIdentifier.firstPathSegment())
	return os.RemoveAll(firstPathSegmentWithRoot)
}

func (s *Storage) ReadFile(fileName string) (io.Reader, int64, error) {
	return s.readIntoFile(fileName)
}

func (s *Storage) ReadFileDecrypted(fileName string, decryptFunc func([]byte, io.Writer, io.Reader) (int64, error), key []byte) (io.Reader, int64, error) {
	file, size, err := s.readIntoFile(fileName)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	// Create a buffer to handle the decrypted data
	var buf bytes.Buffer

	// Decrypt the data directly into the buffer
	_, err = decryptFunc(key, &buf, file)
	if err != nil {
		return nil, 0, err
	}

	return &buf, size, nil
}

// readIntoFile opens a specified file for reading.
// It uses the PathTransformFunc from the Storage configuration to transform the file name.
// The function constructs the full file path by appending the root directory.
// It then opens the file at the specified full path and retrieves the file's size upon successful opening.
func (s *Storage) readIntoFile(fileName string) (io.ReadCloser, int64, error) {
	fileIdentifier := s.Config.PathTranformFunc(fileName)
	fullPathWithRoot := s.prependTheRoot(fileIdentifier.BuildFilePath())

	f, err := os.Open(fullPathWithRoot)
	if err != nil {
		return nil, 0, err
	}

	fileInfo, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}

	return f, fileInfo.Size(), nil
}

// storeToDestinationFile is a helper function that manages file path setup and creation,
// allowing for either plain or encrypted data copying using the provided copyFunc.
func (s *Storage) storeToDestinationFile(fileName string, inputStream io.Reader, copyFunc func(io.Writer, io.Reader) (int64, error)) (int64, error) {
	// Transform and prepare the file path
	fileIdentifier := s.Config.PathTranformFunc(fileName)
	pathNameWithRoot := s.prependTheRoot(fileIdentifier.PathName)

	// Ensure the directory structure exists
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return 0, err
	}

	// Get the full file path
	fullPathWithRoot := s.prependTheRoot(fileIdentifier.BuildFilePath())

	// Create the destination file
	destinationFile, err := os.Create(fullPathWithRoot)
	if err != nil {
		return 0, err
	}
	defer destinationFile.Close()

	// Use the specified copy function to write data to the file
	return copyFunc(destinationFile, inputStream)
}

// StoreFile reads from the input stream and writes unencrypted data to a file.
func (s *Storage) StoreFile(fileName string, inputStream io.Reader) (int64, error) {
	// Use io.Copy for direct data copying
	return s.storeToDestinationFile(fileName, inputStream, func(dst io.Writer, src io.Reader) (int64, error) {
		return io.Copy(dst, src)
	})
}

// StoreFileEncrypted reads from the input stream, encrypts the data using the provided encryptFunc to be more flexible,
// and writes it to a file.
func (s *Storage) StoreFileEncrypted(fileName string, inputStream io.Reader, encryptFunc func([]byte, io.Writer, io.Reader) (int64, error), key []byte) (int64, error) {
	// Use the user-defined encryptFunc for encrypted data copying
	return s.storeToDestinationFile(fileName, inputStream, func(dst io.Writer, src io.Reader) (int64, error) {
		return encryptFunc(key, dst, src)
	})
}
