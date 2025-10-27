package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os/exec"
	"strings"
	"time"
)

var persistentClient *http.Client

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Failed to create cookie jar: %v", err)
	}

	persistentClient = &http.Client{
		Jar:     jar,
		Timeout: 120 * time.Second,
	}

	media := search(SEARCHSTRING)

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

	var subs string
	if !NOSUBS {
		subslink := mediaJSON.Tracks[findLanguage(mediaJSON, LANGUAGE)].File
		subs = fmt.Sprintf(`--sub-file=%s`, subslink)

	}
	if DEBUG {
		fmt.Printf(fmt.Sprintf("Running: mpv %s %s %s", link, title, subs))
	}
	cmd := exec.Command("mpv", link, title, subs)
	mpvErr := cmd.Run()
	if mpvErr != nil {
		panic("Unable to spawn MPV")
	}
}
