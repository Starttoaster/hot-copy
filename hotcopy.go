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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
)

var watch *watcher.Watcher = watcher.New()
var jobQueue []watcher.Event
var decryptDir string = "/data"
var encryptDir string = "/enc-data"
var puid, pgid int

func encryptFile(key []byte, path string, outFilename string, fileMode os.FileMode) {
	plaintext, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	of, err := os.Create(outFilename)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.Chmod(outFilename, fileMode); err != nil {
		log.Fatal(err)
	}
	if err := os.Chown(outFilename, puid, pgid); err != nil {
		log.Fatal(err)
	}
	defer of.Close()

	//Write the original plaintext size into the output file first, encoded in an 8-byte integer.
	origSize := uint64(len(plaintext))
	if err = binary.Write(of, binary.LittleEndian, origSize); err != nil {
		log.Fatal(err)
	}

	//Pad plaintext to a multiple of BlockSize with random padding.
	if len(plaintext)%aes.BlockSize != 0 {
		bytesToPad := aes.BlockSize - (len(plaintext) % aes.BlockSize)
		padding := make([]byte, bytesToPad)
		if _, err := rand.Read(padding); err != nil {
			log.Fatal(err)
		}
		plaintext = append(plaintext, padding...)
	}

	//Generate random IV and write it to the output file.
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		log.Fatal(err)
	}
	if _, err = of.Write(iv); err != nil {
		log.Fatal(err)
	}

	//Ciphertext has the same size as the padded plaintext.
	ciphertext := make([]byte, len(plaintext))

	//Use AES implementation of the cipher.Block interface to encrypt the whole file in CBC mode.
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

func decryptFile(key []byte, path string, outFilename string, fileMode os.FileMode) {
	ciphertext, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	of, err := os.Create(outFilename)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.Chmod(outFilename, fileMode); err != nil {
		log.Fatal(err)
	}
	if err := os.Chown(outFilename, puid, pgid); err != nil {
		log.Fatal(err)
	}
	defer of.Close()

	//cipertext has the original plaintext size in the first 8 bytes, then IV
	//in the next 16 bytes, then the actual ciphertext in the rest of the buffer.
	//Read the original plaintext size, and the IV.
	var origSize uint64
	buf := bytes.NewReader(ciphertext)
	if err = binary.Read(buf, binary.LittleEndian, &origSize); err != nil {
		log.Fatal(err)
	}
	iv := make([]byte, aes.BlockSize)
	if _, err = buf.Read(iv); err != nil {
		log.Fatal(err)
	}

	//The remaining ciphertext has size=paddedSize.
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

//Grabs environment variables for password, uid, and gid
func getEnv() string {
	password := os.Getenv("HC_PASSWORD")
	if password == "" {
		fmt.Println("Password environment variable not set")
		os.Exit(1)
	}
	uid, gid := os.Getenv("PUID"), os.Getenv("PGID")
	var err error
	puid, err = strconv.Atoi(uid)
	if err != nil {
		log.Panicln(err)
	}
	pgid, err = strconv.Atoi(gid)
	if err != nil {
		log.Panicln(err)
	}

	return password
}

//Makes a hashed 32 byte key out of the user-set password
func makeKey(password string) []byte {
	hashkey := sha256.Sum256([]byte(password))
	var key []byte = hashkey[:]
	return key
}

//Helper function to change the base directory path from one to the other
func switchFolder(wholePath string, startDir string, newDir string) string {
	newPath := newDir + strings.TrimPrefix(wholePath, startDir)
	return newPath
}

//Starts the directory watcher
func watchDirs() {
	//Selects the directories to recursively search for updates through
	if err := watch.AddRecursive(decryptDir); err != nil {
		log.Panicln(err)
	}
	if err := watch.AddRecursive(encryptDir); err != nil {
		log.Panicln(err)
	}

	//Logic for what gets pushed into the job queue
	go func() {
		for {
			select {
			case event := <-watch.Event:
				if event.Path != encryptDir && event.Path != decryptDir {
					jobQueue = append(jobQueue, event) //Adds event to job queue
				}
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

//Grabs a job from the queue slice, splits it between which directory it comes from, actions it, and then deletes it
func getEvent(key []byte, testing bool) {
	for {
		if len(jobQueue) > 0 {
			oldestEvent := jobQueue[0]
			if !testing {
				fmt.Println(oldestEvent) //Shows event details as they come up the queue
			}
			
			//'If' runs when event occurs in decryptDir. 'Else' runs when event occurs in encryptDir
			if strings.Contains(oldestEvent.Path, decryptDir) {
				//Removes the 'data-enc/' directory watch path while running events in 'data/'
				watch.RemoveRecursive(encryptDir)

					eventHandler(false, key, oldestEvent, false)

				//Re-adds the 'data-enc/' directory watch path once done
				if err := watch.AddRecursive(encryptDir); err != nil {
					log.Panicln(err)
				}
			} else if strings.Contains(oldestEvent.Path, encryptDir) {
				//Removes the 'data/' directory watch path while running events in 'data-enc/'
				watch.RemoveRecursive(decryptDir)

				eventHandler(true, key, oldestEvent, false)

				//Re-adds the 'data/' directory watch path once done
				if err := watch.AddRecursive(decryptDir); err != nil {
					log.Panicln(err)
				}
			} else if testing {
				eventHandler(true, key, oldestEvent, testing)
				eventHandler(false, key, oldestEvent, testing)
			}

			jobQueue = append(jobQueue[:0], jobQueue[1:]...) //Removes the first (oldest) element of the slice
		} else if !testing && len(jobQueue) == 0 {
			time.Sleep(1 * time.Second)
		} else if testing && len(jobQueue) == 0 {
			break
		}
	}
}

//Gets called by getJob with relevant details
func eventHandler(enc bool, key []byte, event watcher.Event, testing bool) {
	eventPath := event.Path

	if event.Op.String() == "WRITE" || event.Op.String() == "CREATE" {
		eventName := event.Name()
		eventIsDir := event.IsDir()
		eventMode := event.Mode()
		if !testing {
			if eventIsDir && event.Path != decryptDir {
				dirName := switchFolder(eventPath, decryptDir, encryptDir)
				os.MkdirAll(dirName, eventMode)
			} else if eventIsDir && eventPath != encryptDir {
				dirName := switchFolder(eventPath, encryptDir, decryptDir)
				os.MkdirAll(dirName, eventMode)
			} else {
				writeFile(enc, key, eventPath, eventName, eventMode)
			}
		}
	} else if event.Op.String() == "REMOVE" {
		if !testing {
			eventIsDir := event.IsDir()
			deleteFile(enc, eventPath, eventIsDir)
		}
	} else if event.Op.String() == "RENAME" || event.Op.String() == "MOVE" {
		if !testing {
			eventOldPath := event.OldPath
			renameFile(enc, eventPath, eventOldPath)
		}
	}
}

//Runs when eventHandler reaches a RENAME event. Skips actions if file isn't found.
func renameFile(enc bool, eventPath string, eventOldPath string) {
	var oldName string
	var newName string
	if !enc {
		oldName = switchFolder(eventOldPath, decryptDir, encryptDir)
		newName = switchFolder(eventPath, decryptDir, encryptDir)
	} else {
		oldName = switchFolder(eventOldPath, encryptDir, decryptDir)
		newName = switchFolder(eventPath, encryptDir, decryptDir)
	}
	if _, err := os.Stat(oldName); err == nil {
		err := os.Rename(oldName, newName)
		if err != nil {
			log.Fatal(err)
		}
	}
}

//Runs when eventHandler reaches a REMOVE event. Skips actions if file isn't found.
func deleteFile(enc bool, path string, isDir bool) {
	var toDel string
	if !enc {
		toDel = switchFolder(path, decryptDir, encryptDir)
	} else {
		toDel = switchFolder(path, encryptDir, decryptDir)
	}
	if _, err := os.Stat(toDel); err == nil {
		if isDir {
			d, err := os.Open(toDel)
			if err != nil {
				log.Panicln(err)
			}
			names, err := d.Readdirnames(0)
			if err != nil {
				log.Panicln(err)
			}
			d.Close()
			for _, name := range names {
				err = os.RemoveAll(filepath.Join(toDel, name))
				if err != nil {
					log.Panicln(err)
				}
			}
			err = os.Remove(toDel)
		} else {
			err := os.Remove(toDel)
			if err != nil {
				log.Panicln(err)
			}
		}
	}
}

//Runs when eventHandler reaches a Create or Write event
func writeFile(enc bool, key []byte, eventPath string, eventName string, eventMode os.FileMode) {
	if !enc {
		checkDir := strings.TrimSuffix(switchFolder(eventPath, decryptDir, encryptDir), eventName)
		outFilename := switchFolder(eventPath, decryptDir, encryptDir)
		os.MkdirAll(checkDir, eventMode)
		encryptFile(key, eventPath, outFilename, eventMode)
	} else {
		checkDir := strings.TrimSuffix(switchFolder(eventPath, encryptDir, decryptDir), eventName)
		outFilename := switchFolder(eventPath, encryptDir, decryptDir)
		os.MkdirAll(checkDir, eventMode)
		decryptFile(key, eventPath, outFilename, eventMode)
	}
}

func main() {
	//Set up user defined key, and uid/gid variables
	key := makeKey(getEnv())

	go getEvent(key, false) //Start event queue goroutine
	watchDirs()      //Start recursive directory watcher
}
