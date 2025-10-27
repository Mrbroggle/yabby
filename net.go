package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getM3U8(json MediaData) (string, error) {
	for _, element := range json.Sources {
		file := element.File
		if strings.Contains(file, "playlist") {
			return file, nil
		}
	}
	return "", errors.New("no index m3u8")
}

func docGet(uri string) *goquery.Document {
	res := httpGet(uri)
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func httpGet(uri string) *http.Response {
	if DEBUG {
		fmt.Printf("Gettting: %s\n", uri)
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "curl/69.420.0")
	res, err := persistentClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	if DEBUG {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("error reading response body for logging: %v", err)
		}
		fmt.Println(string(bodyBytes))
		res.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	return res
}

func getJSON(uri string, target any) {
	res := httpGet(uri)
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(res.Body); err != nil {
		panic("Error reading embed response body:")
	}

	jsonErr := json.Unmarshal(buf.Bytes(), target)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
}
