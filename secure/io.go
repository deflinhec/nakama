// Copyright 2032 Deflinhec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"go.uber.org/atomic"
)

var encryptionKey *atomic.String

func UseKey(key string) {
	key = fmt.Sprintf("%32s", key)
	if encryptionKey == nil {
		encryptionKey = atomic.NewString(key)
	} else {
		encryptionKey.Store(key)
	}
}

func ReadFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if encryptionKey == nil {
		return content, err
	}

	input := string(content)
	if input == "" {
		return nil, errors.New("empty input")
	}

	if len(input) < aes.BlockSize {
		return nil, errors.New("input too short")
	}

	block, err := aes.NewCipher([]byte(encryptionKey.Load()))
	if err != nil {
		return nil, err
	}

	cipherText := []byte(input)
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}

func WriteFile(path string, content []byte, mode fs.FileMode) error {
	if encryptionKey == nil {
		return os.WriteFile(path, content, os.ModePerm)
	}
	input := string(content)
	// Pad string up to length multiple of 4 if needed.
	if maybePad := len(input) % 4; maybePad != 0 {
		input += strings.Repeat(" ", 4-maybePad)
	}

	block, err := aes.NewCipher([]byte(encryptionKey.Load()))
	if err != nil {
		return err
	}

	cipherText := make([]byte, aes.BlockSize+len(input))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], []byte(input))

	return os.WriteFile(path, cipherText, mode)
}
