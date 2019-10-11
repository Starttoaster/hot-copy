package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var decryptedFile string = "data/test.txt"
var encryptedFile string = "enc-data/test.enc"
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
	os.Setenv("SA_PASSWORD", "testkey")
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
