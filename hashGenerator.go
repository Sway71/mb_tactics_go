package main

import (
	"math/rand"
	"log"
	"encoding/base64"
)

func RandomString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		log.Fatalln(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}