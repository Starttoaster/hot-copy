package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"github.com/radovskyb/watcher"
)

var decryptedFile string = "/data/test.txt"
var encryptedFile string = "/enc-data/test.txt"
var testingText []byte = []byte("testing")

func TestEncryptFile(t *testing.T) {
	testFile, err := os.Create(decryptedFile)
	if err != nil {
		log.Fatal(err)
	}
	defer testFile.Close()
	testFile.Write(testingText)

	key := makeKey(getEnv())
	encryptFile(key, decryptedFile, encryptedFile, 0644)

	//Tests to ensure encrypted file was created
	if _, err := os.Stat(encryptedFile); os.IsNotExist(err) {
		t.Fail()
	}

	//Tests to ensure encrypted file doesn't equal the decrypted one
	encryptedText, err := ioutil.ReadFile(encryptedFile)
	if err != nil {
		log.Fatal(err)
	}
	if string(encryptedText) == string(testingText) {
		t.Fail()
	}

	os.Remove(decryptedFile) //Removes the unencrypted file
}

func TestDecryptFile(t *testing.T) {
	key := makeKey(getEnv())
	decryptFile(key, encryptedFile, decryptedFile, 0644)

	//Tests to ensure decrypted file was created
	if _, err := os.Stat(decryptedFile); os.IsNotExist(err) {
		t.Fail()
	}

	//Tests to ensure decrypted file equals the original text
	decryptedText, err := ioutil.ReadFile(decryptedFile)
	if err != nil {
		log.Fatal(err)
	}
	if string(decryptedText) != string(testingText) {
		t.Fail()
	}

	//Removes both files
	os.Remove(encryptedFile)
	os.Remove(decryptedFile)
}

func TestGetEnv(t *testing.T) {
	//Sets testing variables
	os.Setenv("HC_PASSWORD", "testkey")
	os.Setenv("PUID", "1000")
	os.Setenv("PGID", "1000")

	//Tests using environment variables to make a key
	key := getEnv()
	if key != "testkey" {
		t.Fail()
	}
	//Tests the puid/pgid variable sett
	if puid != 1000 || pgid != 1000 {
		t.Fail()
	}
}

func TestMakeKey(t *testing.T) {
	key := makeKey("testkey")
	//Ensures key created is always 32 bytes in length
	if len(key) != 32 {
		t.Fail()
	}
}

func TestSwitchFolder(t *testing.T) {
	newPath := switchFolder("/oldpath/somedirectory/file", "/oldpath", "/newpath")
	if newPath != "/newpath/somedirectory/file" {
		t.Fail()
	}
}

func TestRenameFile(t *testing.T) {
	//New file names
	newDecryptedFile := "/data/cool.txt"
	newEncryptedFile := "/enc-data/cool.txt"

	//Creating files that will be changed
	testFile, err := os.Create(decryptedFile)
	if err != nil {
		t.Fail()
	}
	testEncFile, err := os.Create(encryptedFile)
	if err != nil {
		t.Fail()
	}
	defer testFile.Close()
	defer testEncFile.Close()

	//Renames file in encrypted directory then tests to make sure it changed
	renameFile(false, newDecryptedFile, decryptedFile)
	if _, err := os.Stat(newEncryptedFile); err != nil {
		t.Fail()
	}

	//Renames file in decrypted directory then tests to make sure it changed
	renameFile(true, newEncryptedFile, encryptedFile)
	if _, err := os.Stat(newDecryptedFile); err != nil {
		t.Fail()
	}
}

func TestDeleteFile(t *testing.T) {
	decryptedFolder := "/data/folder"
	encryptedFolder := "/enc-data/folder"

	//Creating files that will be changed
	testFile, err := os.Create(decryptedFile)
	if err != nil {
		t.Fail()
	}
	testEncFile, err := os.Create(encryptedFile)
	if err != nil {
		t.Fail()
	}
	os.MkdirAll(decryptedFolder, 0644)
	os.MkdirAll(encryptedFolder, 0644)
	defer testFile.Close()
	defer testEncFile.Close()

	//Deletes the files in encrypted directory, and then tests to make sure they were removed
	deleteFile(false, decryptedFile, false)
	deleteFile(false, decryptedFolder, true)
	if _, err := os.Stat(encryptedFile); err == nil {
		t.Fail()
	}
	if _, err := os.Stat(encryptedFolder); err == nil {
		t.Fail()
	}

	//Deletes the files in decrypted directory, and then tests to make sure they were removed
	deleteFile(true, encryptedFile, false)
	deleteFile(true, encryptedFolder, true)
	if _, err := os.Stat(decryptedFile); err == nil {
		t.Fail()
	}
	if _, err := os.Stat(decryptedFolder); err == nil {
		t.Fail()
	}
}

func TestWriteFile(t *testing.T) {
	key := makeKey("testkey")
	decryptedFileName := "test.txt"
	encryptedFileName := "test.txt"

	//Creating decrypted files that will be written into the encrypted directory
	testFile, err := os.Create(decryptedFile)
	if err != nil {
		t.Fail()
	}
	defer testFile.Close()
	//Testing writeFile first run
	writeFile(false, key, decryptedFile, decryptedFileName, 0644)
	if _, err := os.Stat(encryptedFile); err != nil {
		t.Fail()
	}
	os.Remove(decryptedFile) //Getting rid of decryptedFile just to create it again
	//Testing writeFile second run
	writeFile(true, key, encryptedFile, encryptedFileName, 0644)
	if _, err := os.Stat(decryptedFile); err != nil {
		t.Fail()
	}
}

func TestWatchDirs(t *testing.T) {
	//Start recursive directory watcher and trigger all event types
	go watchDirs()
	watch.TriggerEvent(watcher.Create, nil)
	watch.TriggerEvent(watcher.Write, nil)
	watch.TriggerEvent(watcher.Remove, nil)
	watch.TriggerEvent(watcher.Rename, nil)
	watch.TriggerEvent(watcher.Chmod, nil)
	watch.TriggerEvent(watcher.Move, nil)
	watch.Close()
	//Ensure all event types were recognized and added to the queue
	if len(jobQueue) != 6 {
		t.Fail()
	}
}

func TestGetEvent(t *testing.T) {
	key := makeKey("testkey")

	//Makes sure jobQueue has events in it
	if len(jobQueue) == 0 {
		t.Fail()
	}
	//Flush out all events
	getEvent(key, true)
	if len(jobQueue) != 0 {
		t.Fail()
	}
}


