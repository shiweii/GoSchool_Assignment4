package encryptdecrypt

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

/*
	File might remains decrypted if a panic occur,
	this function will encrypt the file.
*/
func CheckEncryption(encryptedFile string, decryptedFile string) {
	_, err := ioutil.ReadFile(encryptedFile)
	if err != nil {
		EncryptFile(decryptedFile, encryptedFile)
	}
}

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
