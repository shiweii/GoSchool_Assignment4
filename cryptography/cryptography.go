// Package cryptography implements cryptography for file encryption and decryption
package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"

	"github.com/shiweii/logger"
	util "github.com/shiweii/utility"
)

// CheckEncryption check if file is encrypted.
// Will perform encryption if file is not encrypted.
func CheckEncryption(encryptedFile string, decryptedFile string) {
	_, err := ioutil.ReadFile(encryptedFile)
	if err != nil {
		EncryptFile(decryptedFile, encryptedFile)
	}
}

// EncryptFile performs AES-256 encryption on file, plaintext file will then be deleted.
func EncryptFile(path string, resultPath string) {
	plaintext, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Println(err)
	}

	key := []byte(util.GetEnvVar("KEY"))

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Error.Println(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error.Println(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Error.Println(err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	err = ioutil.WriteFile(resultPath, ciphertext, 0777)
	if err != nil {
		logger.Error.Println(err)
	}

	// Delete plain text file
	e := os.Remove(path)
	if e != nil {
		logger.Error.Println(err)
	}
}

// DecryptFile performs AES-256 description on file, encrypted file will then be deleted.
func DecryptFile(path string, resultPath string) {
	ciphertext, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Println(err)
	}

	key := []byte(util.GetEnvVar("KEY"))

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Error.Println(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error.Println(err)
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		logger.Error.Println(err)
	}

	err = ioutil.WriteFile(resultPath, plaintext, 0777)
	if err != nil {
		logger.Error.Println(err)
	}

	// Delete encrypted file
	e := os.Remove(path)
	if e != nil {
		logger.Error.Println(err)
	}
}
