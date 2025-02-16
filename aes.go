package aes

import (
	"bytes"
	cryptoaes "crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

var cipherKey []byte

func init() {
	cipherKeyString := os.Getenv("CIPHER_KEY") // Corrected the environment variable name
	if cipherKeyString == "" {
		log.Println("CIPHER_KEY environment variable is not set. Generating a new key.")
		cipherKey = newKey()
	} else {
		// Assuming the key is base64 encoded
		decodedKey, err := base64.StdEncoding.DecodeString(cipherKeyString)
		if err != nil || (len(decodedKey) != 16 && len(decodedKey) != 24 && len(decodedKey) != 32) {
			log.Fatalf("Invalid CIPHER_KEY provided. Please ensure it is base64 encoded and 16, 24, or 32 bytes long after decoding.")
		}
		cipherKey = decodedKey
	}

	log.Printf("Using cipher key of size %d bytes.\n", len(cipherKey))
}

func newKey() []byte {
	// Generate a new random key
	key := make([]byte, 32) // 32 bytes for AES-256
	if _, err := rand.Read(key); err != nil {
		log.Fatal("Failed to generate a new cipher key")
	}
	return key
}
func addBase64Padding(value string) string {
	m := len(value) % 4
	if m != 0 {
		value += strings.Repeat("=", 4-m)
	}

	return value
}

func removeBase64Padding(value string) string {
	return strings.Replace(value, "=", "", -1)
}

func pad(src []byte) []byte {
	padding := cryptoaes.BlockSize - len(src)%cryptoaes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

func Encrypt(text string) (string, error) {
	block, err := cryptoaes.NewCipher(cipherKey)
	if err != nil {
		return "", err
	}

	msg := pad([]byte(text))
	ciphertext := make([]byte, cryptoaes.BlockSize+len(msg))
	iv := ciphertext[:cryptoaes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[cryptoaes.BlockSize:], []byte(msg))
	finalMsg := removeBase64Padding(base64.URLEncoding.EncodeToString(ciphertext))
	return finalMsg, nil
}

func Decrypt(text string) (string, error) {
	block, err := cryptoaes.NewCipher(cipherKey)
	if err != nil {
		return "", err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(addBase64Padding(text))
	if err != nil {
		return "", err
	}

	if (len(decodedMsg) % cryptoaes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multipe of decoded message length")
	}

	if len(decodedMsg) == 0 {
		return "", nil
	}

	iv := decodedMsg[:cryptoaes.BlockSize]
	msg := decodedMsg[cryptoaes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := unpad(msg)
	if err != nil {
		return "", err
	}

	return string(unpadMsg), nil
}
