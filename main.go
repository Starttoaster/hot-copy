package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var decryptDir string = "data/"
var encryptDir string = "inside/"

func encryptFile(key []byte, oldPath string, newPath string, filename string) {
	outFilename := newPath + strings.TrimPrefix(filename, oldPath)

	plaintext, err := ioutil.ReadFile(oldPath + filename)
	if err != nil {
		log.Fatal(err)
	}

	of, err := os.Create(outFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer of.Close()

	// Write the original plaintext size into the output file first, encoded in a 8-byte integer.
	origSize := uint64(len(plaintext))
	if err = binary.Write(of, binary.LittleEndian, origSize); err != nil {
		log.Fatal(err)
	}

	// Pad plaintext to a multiple of BlockSize with random padding.
	if len(plaintext)%aes.BlockSize != 0 {
		bytesToPad := aes.BlockSize - (len(plaintext) % aes.BlockSize)
		padding := make([]byte, bytesToPad)
		if _, err := rand.Read(padding); err != nil {
			log.Fatal(err)
		}
		plaintext = append(plaintext, padding...)
	}

	// Generate random IV and write it to the output file.
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		log.Fatal(err)
	}
	if _, err = of.Write(iv); err != nil {
		log.Fatal(err)
	}

	// Ciphertext has the same size as the padded plaintext.
	ciphertext := make([]byte, len(plaintext))

	// Use AES implementation of the cipher.Block interface to encrypt the whole file in CBC mode.
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	if _, err = of.Write(ciphertext); err != nil {
		log.Fatal(err)
	}
}

func decryptFile(key []byte, oldPath string, newPath string, filename string) {
	outFilename := newPath + strings.TrimPrefix(filename, oldPath)

	ciphertext, err := ioutil.ReadFile(oldPath + filename)
	if err != nil {
		log.Fatal(err)
	}

	of, err := os.Create(outFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer of.Close()

	// cipertext has the original plaintext size in the first 8 bytes, then IV
	// in the next 16 bytes, then the actual ciphertext in the rest of the buffer.
	// Read the original plaintext size, and the IV.
	var origSize uint64
	buf := bytes.NewReader(ciphertext)
	if err = binary.Read(buf, binary.LittleEndian, &origSize); err != nil {
		log.Fatal(err)
	}
	iv := make([]byte, aes.BlockSize)
	if _, err = buf.Read(iv); err != nil {
		log.Fatal(err)
	}

	// The remaining ciphertext has size=paddedSize.
	paddedSize := len(ciphertext) - 8 - aes.BlockSize
	if paddedSize%aes.BlockSize != 0 {
		log.Fatal(fmt.Errorf("ERROR: want padded plaintext size to be aligned to block size"))
	}
	plaintext := make([]byte, paddedSize)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext[8+aes.BlockSize:])

	if _, err := of.Write(plaintext[:origSize]); err != nil {
		log.Fatal(err)
	}
}

//Function reads through the selected base directory (encrypted or no) and replicates it in the other directory
func searchDirs(searchDir string, createDir string, enc bool, key []byte) {
	files, err := ioutil.ReadDir(searchDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if f.IsDir() == true {
			newSearchDir := searchDir + f.Name() + "/"
			newCreateDir := createDir + f.Name() + "/"
			os.MkdirAll(newCreateDir, f.Mode())
			searchDirs(newSearchDir, newCreateDir, enc, key)
		} else {
			if enc {
				encryptFile(key, searchDir, createDir, f.Name())
			} else {
				decryptFile(key, searchDir, createDir, f.Name())
			}
		}
	}
}

//Requires linux environment variable "SA_PASSWORD" be set
func getPass() string {
	password := os.Getenv("SA_PASSWORD")
	if password == "" {
		fmt.Println("Password environment variable not set")
		os.Exit(1)
	}
	return password
}

//Makes a hashed 32 byte key out of the user-set password
func makeKey(pass string) []byte {
	hashkey := sha256.Sum256([]byte(pass))
	var key []byte = hashkey[:]
	return key
}

//Compares the modification times of both main directories and runs the searchDirs function if applicable
func compareDirs(key []byte) {
	for {
		infoDecryptDir, err := os.Stat(decryptDir)
		if err != nil {
			log.Fatal(err)
		}
		infoEncryptDir, err := os.Stat(encryptDir)
		if err != nil {
			log.Fatal(err)
		}
		decryptDirTime := infoDecryptDir.ModTime()
		encryptDirTime := infoEncryptDir.ModTime()

		diff := decryptDirTime.Sub(encryptDirTime)

		if diff == (time.Duration(0) * time.Second) {
			fmt.Println("Doing nothing this run...")
			break
		} else if diff < (time.Duration(0) * time.Second) {
			searchDirs(encryptDir, decryptDir, false, key)
		} else if diff > (time.Duration(0) * time.Second) {
			searchDirs(decryptDir, encryptDir, true, key)
		}
		time.Sleep(10 * time.Second)
	}
}

func main() {
	password := getPass()
	key := makeKey(password)

	compareDirs(key)
}
