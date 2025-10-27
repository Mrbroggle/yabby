package main

import "net/http"

var (
	BASE             = "https://flixhq.to"
	PROVIDER         = "Vidcloud"
	DEBUG            = false
	DECODEURL        = "https://dec.eatmynerds.live"
	RICHPRESENCE     = false
	QUALITY          = "720"
	persistentClient *http.Client
)
