package main

import (
	"fmt"
	"time"

	"github.com/hugolgst/rich-go/client"
)

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
