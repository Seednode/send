/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

const (
	ReleaseVersion string = "0.6.1"
)

var Count uint32
var Domain string
var Length uint16
var Port uint16
var Randomize bool
var Scheme string
var URI string
var Verbose bool

var rootCmd = &cobra.Command{
	Use:   "send [file]...",
	Short: "Generates a one-off download link for one or more specified file(s).",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && !isFromPipe() {
			err := errors.New("no file(s) specified and no data received from stdin")
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ServePage(args)
	},
}

func Execute() {
	rootCmd.Execute()
}

func init() {
	rootCmd.Flags().Uint32VarP(&Count, "count", "c", 0, "number of times to serve the file(s)")
	rootCmd.Flags().StringVarP(&Domain, "domain", "d", "localhost", "domain to use in returned urls")
	rootCmd.Flags().Uint16VarP(&Length, "length", "l", 6, "length of url slug and obfuscated filename(s)")
	rootCmd.Flags().Uint16VarP(&Port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVarP(&Randomize, "randomize", "r", false, "randomize filename(s)")
	rootCmd.Flags().StringVarP(&Scheme, "scheme", "s", "http", "scheme to use in returned urls")
	rootCmd.Flags().StringVarP(&URI, "uri", "u", "", "full uri (overrides domain, scheme, and port)")
	rootCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "log accessed files to stdout")

	rootCmd.Flags().SetInterspersed(true)

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SilenceErrors = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	rootCmd.SetVersionTemplate("send v{{.Version}}\n")
	rootCmd.Version = ReleaseVersion
}
