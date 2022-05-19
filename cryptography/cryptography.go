// Package cryptography implements cryptographic functions
// and file encryption or decryption.
package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"

	"github.com/shiweii/logger"
)

// EncryptFile performs AES-256 encryption on file, plaintext file will then be deleted.
func EncryptFile(envKey, path, resultPath string) {
	plaintext, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Println(err)
	}

	key := []byte(envKey)

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
func DecryptFile(envKey, path, resultPath string) {
	ciphertext, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Println(err)
	}

	key := []byte(envKey)

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

// ComputeSHA512 computes the SHA512 checksum of a given file.
func ComputeSHA512(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logger.Error.Println(err)
		}
	}(f)

	harsher := sha512.New()
	if _, err := io.Copy(harsher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(harsher.Sum(nil)), nil
}
