package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var encryptedFile string = "test.enc"
var decryptedFile string = "test.txt"
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

func TestMakeKey(t *testing.T) {
	key := makeKey("testkey")
	//Ensures key created is always 32 bytes in length
	if len(key) != 32 {
		t.Fail()
	}
}
