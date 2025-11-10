/*
Copyright © 2025 Seednode <seednode@seedno.de>
*/

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// Version number for built binaries and Docker image releases
	ReleaseVersion string = "3.1.0"
)

var (
	// The IP address on which send will listen
	Bind string

	// The number of times to serve selected file(s) before shutting down
	Count int

	// Exit on error, instead of just printing the error
	ErrorExit bool

	// The length of randomly generated slugs and filenames
	Length int

	// The port on which send will listen
	Port int

	// Register http/pprof handlers
	Profile bool

	// Randomize filenames in URLs
	Randomize bool

	// Scheme to use in generated URLs
	Scheme string

	// The length of time after which send will shut down
	Timeout time.Duration

	// How often to display remaining time before shutdown, when timeout is enabled
	TimeoutInterval time.Duration

	// TLS certificate and key for HTTPS connections
	TLSCert string
	TLSKey  string

	// Value to be used instead of http://<bind>:<port> in returned links
	URL string
)

func main() {
	cmd := &cobra.Command{
		Use:   "send [file]...",
		Short: "Generates a one-off download link for one or more specified files.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initializeConfig(cmd)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case TLSCert == "" && TLSKey != "" || TLSCert != "" && TLSKey == "":
				return ErrInvalidTLSConfig
			case Count < 0:
				return ErrInvalidCount
			case Length < 0:
				return ErrInvalidLength
			case Port < 1 || Port > 65535:
				return ErrInvalidPort
			case len(args) == 0 && !isFromPipe():
				return ErrNoFile
			}

			if TLSCert != "" && TLSKey != "" && Scheme == "http" {
				Scheme = "https"
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ServePage(args)
		},
	}

	cmd.Flags().StringVarP(&Bind, "bind", "b", "0.0.0.0", "address to bind to")
	cmd.Flags().IntVarP(&Count, "count", "c", 0, "number of times to serve files")
	cmd.Flags().BoolVarP(&ErrorExit, "exit", "e", false, "shut down webserver on error, instead of just printing error")
	cmd.Flags().IntVarP(&Length, "length", "l", 6, "length of url slug and obfuscated filenames")
	cmd.Flags().IntVarP(&Port, "port", "p", 8080, "port to listen on")
	cmd.Flags().BoolVar(&Profile, "profile", false, "register net/http/pprof handlers")
	cmd.Flags().BoolVarP(&Randomize, "randomize", "r", false, "randomize filenames")
	cmd.Flags().StringVarP(&Scheme, "scheme", "s", "http", "scheme to use in returned URLs")
	cmd.Flags().DurationVarP(&Timeout, "timeout", "t", 0, "shutdown after this length of time")
	cmd.Flags().DurationVarP(&TimeoutInterval, "interval", "i", time.Minute, "display remaining time in timeout at this interval")
	cmd.Flags().StringVar(&TLSCert, "tls-cert", "", "path to TLS certificate")
	cmd.Flags().StringVar(&TLSKey, "tls-key", "", "path to TLS keyfile")
	cmd.Flags().StringVarP(&URL, "url", "u", "", "use this value instead of <scheme>://<bind>:<port> in returned URLs")

	cmd.CompletionOptions.HiddenDefaultCmd = true

	cmd.Flags().SetInterspersed(true)

	cmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	cmd.SetVersionTemplate("send v{{.Version}}\n")

	cmd.SilenceErrors = true

	cmd.SilenceUsage = true

	cmd.Version = ReleaseVersion

	log.SetFlags(0)

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func initializeConfig(cmd *cobra.Command) {
	v := viper.New()

	v.SetEnvPrefix("send")

	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	v.AutomaticEnv()

	bindFlags(cmd, v)
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := strings.ReplaceAll(f.Name, "-", "_")

		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
