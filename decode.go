package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type EmbedData struct {
	Link string `json:"link"`
}

type ChallengeData struct {
	Payload    string `json:"payload"`
	Signature  string `json:"signature"`
	Difficulty int    `json:"difficulty"`
}

func extractFromEmbed(embedLink string) MediaData {
	var ChallengeData ChallengeData
	getJSON(fmt.Sprintf("%s/challenge", DECODEURL), &ChallengeData)

	var MediaData MediaData

	nonce := solvePow(ChallengeData.Payload, ChallengeData.Difficulty)
	getJSON(fmt.Sprintf("%s/?url=%s&_debug=true&payload=%s&signature=%s&nonce=%s", DECODEURL, embedLink, ChallengeData.Payload, ChallengeData.Signature, nonce), &MediaData)
	return MediaData
}

func solvePow(payload string, difficulty int) string {
	parts := strings.Split(payload, ".")
	challenge := parts[0]

	prefix := strings.Repeat("0", difficulty)

	nonce := 0
	startTime := time.Now()

	fmt.Printf("Solving PoW challenge (Difficulty %d): %s...\n", difficulty, challenge)

	for {
		text := []byte(fmt.Sprintf("%s%d", challenge, nonce))

		hashBytes := sha256.Sum256(text)
		hashVal := hex.EncodeToString(hashBytes[:])

		if strings.HasPrefix(hashVal, prefix) {
			elapsed := time.Since(startTime).Seconds()
			fmt.Printf("PoW Solved. Nonce: %d, Hash: %s, Time: %.4fs\n", nonce, hashVal, elapsed)
			return strconv.Itoa(nonce)
		}

		nonce++
	}
}
