package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Source struct {
	File string `json:"file"`
}

type Track struct {
	File    string `json:"file"`
	Label   string `json:"label"`
	Kind    string `json:"kind"`
	Default bool   `json:"default"`
}

type MediaData struct {
	Sources []Source `json:"sources"`
	Tracks  []Track  `json:"tracks"`
}

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
