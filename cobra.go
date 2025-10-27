package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yabby [movie]",
	Short: "Lobster but not in shell",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if there are any arguments
		if len(args) > 0 {
			SEARCHSTRING = strings.Join(args, "-")
		} else {
			fmt.Print("Enter movie to search: ")
			fmt.Scanln(&SEARCHSTRING) // Use Print to prompt the user
		}
		SEARCHSTRING = strings.ReplaceAll(SEARCHSTRING, " ", "-")

		fmt.Println("Search String:", SEARCHSTRING)
	},
}

var (
	DEBUG        bool
	RICHPRESENCE bool
	NOSUBS       bool
	BASE         string
	PROVIDER     string
	QUALITY      string
	LANGUAGE     string
	DECODEURL    string
	SEARCHSTRING string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&DEBUG, "debug", "X", false, "Debug output")
	rootCmd.PersistentFlags().BoolVarP(&RICHPRESENCE, "rpc", "r", false, "Enabled discord RPC")
	rootCmd.PersistentFlags().BoolVarP(&NOSUBS, "no-subs", "n", false, "Disables subs")
	rootCmd.PersistentFlags().StringVarP(&BASE, "base", "b", "https://flixhq.to", "Specify the base url")
	rootCmd.PersistentFlags().StringVarP(&PROVIDER, "provider", "p", "Vidcloud", "Specify the provider")
	rootCmd.PersistentFlags().StringVarP(&QUALITY, "quality", "q", "1080", "Specify the video quality")
	rootCmd.PersistentFlags().StringVarP(&LANGUAGE, "language", "l", "English", "Specify the subtitle language")
	rootCmd.PersistentFlags().StringVarP(&DECODEURL, "decoder", "d", "https://dec.eatmynerds.live", "Specify the decode api")
}
