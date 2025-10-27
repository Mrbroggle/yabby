package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/koki-develop/go-fzf"
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
