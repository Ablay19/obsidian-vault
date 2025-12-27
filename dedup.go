package main

import (
"crypto/sha256"
"encoding/hex"
"io"
"os"
)

var processedHashes = make(map[string]string)

func getFileHash(filePath string) (string, error) {
f, err := os.Open(filePath)
if err != nil {
return "", err
}
defer f.Close()
h := sha256.New()
io.Copy(h, f)
return hex.EncodeToString(h.Sum(nil)), nil
}

func isDuplicate(filePath string) bool {
hash, err := getFileHash(filePath)
if err != nil {
return false
}
if _, exists := processedHashes[hash]; exists {
return true
}
processedHashes[hash] = filePath
return false
}
