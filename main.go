package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Failed to create cookie jar: %v", err)
	}

	persistentClient = &http.Client{
		Jar:     jar,
		Timeout: 120 * time.Second,
	}

	media := search(strings.ReplaceAll(os.Args[len(os.Args)-1], " ", "-"))
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

	var EmbedData EmbedData
	getJSON(fmt.Sprintf("%s/ajax/episode/sources/%s", BASE, episode.id), &EmbedData)

	if DEBUG {
		fmt.Println(EmbedData.Link)
	}

	mediaJSON := extractFromEmbed(EmbedData.Link)

	if DEBUG {
		fmt.Println(mediaJSON)
	}

	m3u8Link, err := getM3U8(mediaJSON)

	link := strings.ReplaceAll(m3u8Link, "/playlist.m3u8", fmt.Sprintf("/%s/index.m3u8", QUALITY))

	if err != nil {
		panic(err)
	}

	if RICHPRESENCE {
		RPC(media)
	}

	title := fmt.Sprintf(`--force-media-title=%s`, episode.name)
	cmd := exec.Command("mpv", link, title)
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
