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

	"github.com/radovskyb/watcher"
)

var decryptDir string = "/data"
var encryptDir string = "/inside"
var watch *watcher.Watcher = watcher.New()
var jobQueue []watcher.Event

func encryptFile(key []byte, path string) {
	outFilename := switchFolder(path, decryptDir, encryptDir)

	plaintext, err := ioutil.ReadFile(path)
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

func decryptFile(key []byte, path string) {
	outFilename := switchFolder(path, encryptDir, decryptDir)

	ciphertext, err := ioutil.ReadFile(path)
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

//Helper function to change the base directory path from one to the other
func switchFolder(wholePath string, startDir string, newDir string) string {
	newPath := newDir + strings.TrimPrefix(wholePath, startDir)
	return newPath
}

func watchDirs() {
	//Selects the directories to recursively search for updates through
	if err := watch.AddRecursive(decryptDir); err != nil {
		log.Panicln(err)
	}
	if err := watch.AddRecursive(encryptDir); err != nil {
		log.Panicln(err)
	}

	//Main logic for handling events
	go func() {
		for {
			select {
			case event := <-watch.Event:
				fmt.Println(event) //Shows event details as they occur
				jobQueue = append(jobQueue, event) //Adds event to job queue
			case err := <-watch.Error:
				log.Panicln(err)
			case <-watch.Closed:
				return
			}
		}
	}()
	//Starts the watcher
	if err := watch.Start(time.Millisecond * 100); err != nil {
		log.Panicln(err)
	}
}

func getJob(key []byte) {
	for {
		if len(jobQueue) > 0 {
			oldestEvent := jobQueue[0]

			//'If' runs when event occurs in decryptDir. 'Else' runs when event occurs in encryptDir
			if strings.Contains(oldestEvent.Path, decryptDir) {
				//Removes the 'inside/' directory watch path while running events in 'data/'
				watch.RemoveRecursive(encryptDir)
		
				eventHandler(true, key, oldestEvent)
			
				//Re-adds the 'inside/' directory watch path once done
				if err := watch.AddRecursive(encryptDir); err != nil {
					log.Panicln(err)
				}
			} else {
				//Removes the 'data/' directory watch path while running events in 'inside/'
				watch.RemoveRecursive(encryptDir)
		
				eventHandler(false, key, oldestEvent)
		
				//Re-adds the 'data/' directory watch path once done
				if err := watch.AddRecursive(decryptDir); err != nil {
					log.Panicln(err)
				}
			}
		
			removeOldEvent() //Once done with oldestEvent, removes it from the job queue, and shifts all elements to the left by one
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

//Removes the first (oldest) element of the slice
func removeOldEvent() {
	jobQueue = jobQueue[1:]
}

func eventHandler(enc bool, key []byte, event watcher.Event) {
	if enc {
		//'If' runs when event occurs in decryptDir.
		if event.IsDir() == true && event.Path != decryptDir {
			dirName := switchFolder(event.Path, decryptDir, encryptDir)
			os.MkdirAll(dirName, event.Mode())
		} else if event.IsDir() == true {

		} else {
			if event.Op.String() == "WRITE" || event.Op.String() == "CREATE" {
				writeFile(true, key, &event)
			} else if event.Op.String() == "REMOVE" {
				deleteFile(true, &event)
			}
		}

	} else {
		//'Else' runs when event occurs in encryptDir
		if event.IsDir() == true && event.Path != encryptDir {
			dirName := switchFolder(event.Path, encryptDir, decryptDir)
			os.MkdirAll(dirName, event.Mode())
		} else if event.IsDir() == true {

		} else {
			if event.Op.String() == "WRITE" || event.Op.String() == "CREATE" {
				writeFile(false, key, &event)
			} else if event.Op.String() == "REMOVE" {
				deleteFile(false, &event)
			}
		}
	}
}

//Runs when eventHandler reaches a REMOVE event
func deleteFile(enc bool, event *watcher.Event) {
	var toDel string
	if enc {
		toDel = switchFolder(event.Path, decryptDir, encryptDir)
	} else {
		toDel = switchFolder(event.Path, encryptDir, decryptDir)
	}
	if _, err := os.Stat(toDel); err == nil {
		err := os.Remove(toDel)
		if err != nil {
			log.Panicln(err)
		}
	} else {
		fmt.Println("File did not exist to delete.")
	}
}

//Runs when eventHandler reaches a Create or Write event
func writeFile(enc bool, key []byte, event *watcher.Event) {
	if enc {
		encryptFile(key, event.Path)
	} else {
		decryptFile(key, event.Path)
	}
}

func main() {
	//Set up user defined key
	password := getPass()
	key := makeKey(password)

	go getJob(key)
	watchDirs()
}