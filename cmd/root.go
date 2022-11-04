/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
	Use:   "send [file]",
	Short: "Generates a one-off download link for a specified file.",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
			return err
		}

		if len(args) == 0 && !isFromPipe() {
			err := errors.New("no file specified and no data received from stdin")
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := ServePage(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func Execute() {
	rootCmd.Execute()
}

func init() {
	rootCmd.Flags().Uint32VarP(&Count, "count", "c", 0, "number of times to serve the file")
	rootCmd.Flags().StringVarP(&Domain, "domain", "d", "localhost", "domain to use in returned urls")
	rootCmd.Flags().Uint16VarP(&Length, "length", "l", 6, "length of url slug (and optionally obfuscated filename)")
	rootCmd.Flags().Uint16VarP(&Port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVarP(&Randomize, "randomize", "r", false, "randomize filename")
	rootCmd.Flags().StringVarP(&Scheme, "scheme", "s", "http", "scheme to use in returned urls")
	rootCmd.Flags().StringVarP(&URI, "uri", "u", "", "full uri (overrides domain, scheme, and port)")
	rootCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "log accessed files to stdout")
	rootCmd.Flags().SetInterspersed(true)
}

func isFromPipe() bool {
	f, _ := os.Stdin.Stat()

	if (f.Mode() & os.ModeCharDevice) == 0 {
		return true
	} else {
		return false
	}
}
