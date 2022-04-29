package util

import (
	"crypto/rand"
	"log"
	"math/big"
)

func GenerateSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatal("Could not generate certificate serial number: ", err)
	}

	return serialNumber
}
