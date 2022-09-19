/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var Count uint64
var Domain string
var Length int
var Port int
var Randomize bool
var Scheme string
var Verbose bool

var rootCmd = &cobra.Command{
	Use:   "send <file>",
	Short: "Generates a one-off download link for a specified file.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ServePage(args)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().Uint64VarP(&Count, "count", "c", 0, "number of times to serve the file")
	rootCmd.Flags().StringVarP(&Domain, "domain", "d", "localhost", "domain to use in returned urls")
	rootCmd.Flags().IntVarP(&Length, "length", "l", 6, "length of url slug (and optionally obfuscated filename")
	rootCmd.Flags().IntVarP(&Port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVarP(&Randomize, "randomize", "r", false, "randomize filename")
	rootCmd.Flags().StringVarP(&Scheme, "scheme", "s", "http", "scheme to use in returned urls")
	rootCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "log accessed files to stdout")
	rootCmd.Flags().SetInterspersed(true)
}
