package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/hugolgst/rich-go/client"
	"github.com/koki-develop/go-fzf"
)

var (
	BASE         = "https://flixhq.to"
	PROVIDER     = "Vidcloud"
	DEBUG        = false
	DECODEURL    = "https://dec.eatmynerds.live/?url="
	RICHPRESENCE = false
)

type media struct {
	name string
	id   string
	img  string
	tv   bool
	year string
}
type season struct {
	name string
	id   string
}

type episode struct {
	name string
	id   string
}

func main() {
	media := search(os.Args[len(os.Args)-1])
	flags()

	if RICHPRESENCE {
		RPCSetup()
	}

	if DEBUG {
		fmt.Println(media)
	}
	var episode episode
	if media.tv {
		season := chooseSeason(media)
		if DEBUG {
			fmt.Println(season)
		}
		episode = chooseEpisode(season)
	} else {
		episode = getMovieEpisode(media)
	}
	if DEBUG {
		fmt.Println(episode)
	}

	embedLink := getEmbed(episode)

	if DEBUG {
		fmt.Println(embedLink)
	}

	mediaJSON := extractFromEmbed(embedLink)

	if DEBUG {
		fmt.Println(mediaJSON)
	}

	m3u8Link, err := getM3U8(mediaJSON)
	if err != nil {
		panic(err)
	}

	if RICHPRESENCE {
		RPC(media)
	}

	title := fmt.Sprintf(`--force-media-title=%s`, episode.name)
	cmd := exec.Command("mpv", m3u8Link, title)
	mpvErr := cmd.Run()
	if mpvErr != nil {
		panic("Unable to spawn MPV")
	}
}

func flags() {
	flag.BoolVar(&DEBUG, "X", false, "Enables debug output")
	flag.BoolVar(&RICHPRESENCE, "R", false, "Enables discord rich presence")

	flag.Parse()
}

func search(input string) media {
	doc := docGet(fmt.Sprintf("%s/search/%s", BASE, input))

	var medias []media

	doc.Find("div.flw-item").Each(func(i int, s *goquery.Selection) {
		imageTag := s.Find("img.film-poster-img.lazyload")
		imgURL, existsImg := imageTag.Attr("data-src")

		linkTag := s.Find("a.film-poster-ahref")
		name, existsName := linkTag.Attr("title")
		href, existsHref := linkTag.Attr("href")

		typeTag := s.Find("span.fdi-type")
		mediaType := typeTag.Text()
		isTV := (strings.TrimSpace(mediaType) == "TV")

		yearTag := s.Find("div.fd-infor span.fdi-item").First()
		year := yearTag.Text()

		mediaID := ""
		if existsHref {
			parts := strings.Split(href, "-")
			if len(parts) > 0 {
				mediaID = parts[len(parts)-1]
			}
		}

		if existsName && existsImg && existsHref && mediaID != "" {
			newMedia := media{
				strings.TrimSpace(name),
				mediaID,
				strings.TrimSpace(imgURL),
				isTV,
				strings.TrimSpace(year),
			}
			medias = append(medias, newMedia)
		}
	})

	if len(medias) == 0 {
		panic("No results found.")
	}

	f, err := fzf.New(fzf.WithCaseSensitive(true))
	if err != nil {
		log.Fatal(err)
	}

	idxs, err := f.Find(medias, func(i int) string {
		// Use the year/season field in the display for fzf
		return fmt.Sprintf("%s (%s)", medias[i].name, medias[i].year)
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(idxs) == 0 {
		panic("Selection cancelled.")
	}

	return medias[idxs[0]]
}

func chooseSeason(media media) season {
	doc := docGet(fmt.Sprintf("%s/ajax/v2/tv/seasons/%s", BASE, media.id))

	var seasons []season

	doc.Find("a.ss-item").Each(func(i int, s *goquery.Selection) {
		dataID, existsID := s.Attr("data-id")
		seasonTitle := s.Text()
		if existsID {
			seasons = append(seasons, season{
				seasonTitle,
				dataID,
			})
		}
	})

	f, err := fzf.New(fzf.WithCaseSensitive(true))
	if err != nil {
		log.Fatal(err)
	}

	idxs, err := f.Find(seasons, func(i int) string { return seasons[i].name })
	if err != nil {
		log.Fatal(err)
	}

	return seasons[idxs[0]]
}

func chooseEpisode(season season) episode {
	doc := docGet(fmt.Sprintf("%s/ajax/v2/season/episodes/%s", BASE, season.id))

	var episodes []episode

	doc.Find("a.eps-item").Each(func(i int, s *goquery.Selection) {
		dataID, existsID := s.Attr("data-id")
		episodeTitle, existsTitle := s.Attr("title")
		if existsID && existsTitle {
			episodes = append(episodes, episode{
				episodeTitle,
				dataID,
			})
		}
	})

	f, err := fzf.New(fzf.WithCaseSensitive(true))
	if err != nil {
		log.Fatal(err)
	}

	idxs, err := f.Find(episodes, func(i int) string { return episodes[i].name })
	if err != nil {
		log.Fatal(err)
	}

	return getEpisodeID(episodes[idxs[0]])
}

func getEpisodeID(ep episode) episode {
	doc := docGet(fmt.Sprintf("%s/ajax/v2/episode/servers/%s", BASE, ep.id))

	selection := doc.Find(fmt.Sprintf("a[title='Server %s']", PROVIDER))
	dataID, exists := selection.Attr("data-id")
	if !exists {
		panic("No episode ID")
	}

	return episode{
		ep.name,
		dataID,
	}
}

func getMovieEpisode(media media) episode {
	doc := docGet(fmt.Sprintf("%s/ajax/movie/episodes/%s", BASE, media.id))

	selection := doc.Find(fmt.Sprintf("a[title='%s']", PROVIDER))
	dataID, exists := selection.Attr("data-linkid")
	if !exists {
		panic("No episode ID")
	}

	return episode{
		media.name,
		dataID,
	}
}

type EmbedData struct {
	Link string `json:"link"`
}

func getEmbed(episode episode) string {
	res := httpGet(fmt.Sprintf("%s/ajax/episode/sources/%s", BASE, episode.id))
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(res.Body); err != nil {
		panic("Error reading embed response body:")
	}

	var data EmbedData

	jsonErr := json.Unmarshal(buf.Bytes(), &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return data.Link
}

type Source struct {
	File string `json:"file"`
}

// Track represents an element in the "tracks" array
type Track struct {
	File    string `json:"file"`
	Label   string `json:"label"`
	Kind    string `json:"kind"`
	Default bool   `json:"default"`
}

// MediaData represents the top-level structure of the JSON
type MediaData struct {
	Sources []Source `json:"sources"`
	Tracks  []Track  `json:"tracks"`
}

func extractFromEmbed(embedLink string) MediaData {
	res := httpGet(fmt.Sprintf("%s%s", DECODEURL, embedLink))
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(res.Body); err != nil {
		panic("Error reading embed response body:")
	}

	var data MediaData

	jsonErr := json.Unmarshal(buf.Bytes(), &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return data
}

func getM3U8(json MediaData) (string, error) {
	for _, element := range json.Sources {
		file := element.File
		if strings.Contains(file, "index") {
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
	res, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	return res
}

func RPCSetup() {
	if DEBUG {
		fmt.Println("Setting up RPC")
	}
	err := client.Login("1239340948048187472")
	if err != nil {
		panic(err)
	}
}

func RPC(media media) {
	if DEBUG {
		fmt.Println("Running RPC")
	}
	now := time.Now()
	err := client.SetActivity(client.Activity{
		State:   "Watching",
		Details: media.name,
		Timestamps: &client.Timestamps{
			Start: &now,
		},
	})
	if err != nil {
		panic(err)
	}
}
