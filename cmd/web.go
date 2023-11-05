/*
Copyright © 2023 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	LOGDATE       string = "2006-01-02T15:04:05.000000000-07:00"
	letterBytes          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits        = 6
	letterIdxMask        = 1<<letterIdxBits - 1
	letterIdxMax         = 63 / letterIdxBits
)

type Limits struct {
	channel chan bool
	counter *uint32
}

func isFromPipe() bool {
	f, _ := os.Stdin.Stat()

	if (f.Mode() & os.ModeCharDevice) == 0 {
		return true
	} else {
		return false
	}
}

func generateRandomString(length uint16) string {
	var src = rand.NewSource(time.Now().UnixNano())

	n := int(length)

	builder := strings.Builder{}
	builder.Grow(n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			builder.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return builder.String()
}

func initializeLimits() *Limits {
	channel := make(chan bool, 1)
	var counter uint32

	return &Limits{
		channel: channel,
		counter: &counter,
	}
}

func updateCounter(limits *Limits) {
	atomic.AddUint32(limits.counter, 1)
	counter := atomic.LoadUint32(limits.counter)
	if counter >= Count && Count != 0 {
		defer func() {
			limits.channel <- true
		}()
	}
}

func readStdin() ([]byte, error) {
	var response []byte

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		response = append(response, scanner.Bytes()...)
		response = append(response, "\n"...)
		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
	}

	return response, nil
}

func readFile(path string) ([]byte, error) {
	response, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func serveResponse(w http.ResponseWriter, r http.Request, response []byte, filename string, limits *Limits) error {
	updateCounter(limits)

	var startTime time.Time
	if Verbose {
		startTime = time.Now()
		fmt.Printf("%v | %v requested %v", startTime.Format(LOGDATE), r.RemoteAddr, filename)
	}

	w.Write(response)

	if Verbose {
		fmt.Printf(" (Finished in %v)\n", time.Since(startTime).Round(time.Microsecond))
	}

	return nil
}

func serveResponseHandler(response []byte, filename string, limits *Limits) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := serveResponse(w, *r, response, filename, limits)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func registerHandler(path, slug string, limits *Limits) error {
	var filename string
	switch {
	case Randomize || path == "":
		filename = generateRandomString(Length)
	default:
		filename = filepath.Base(path)
	}

	var response []byte
	var err error = nil

	if path == "" {
		response, err = readStdin()
		if err != nil {
			return err
		}
	} else {
		response, err = readFile(path)
		if err != nil {
			return err
		}
	}

	var url string
	switch URI {
	case "":
		url = fmt.Sprintf("%v://%v:%v/%v/%v", Scheme, Domain, Port, slug, filename)
	default:
		url = fmt.Sprintf("%v/%v/%v", URI, slug, filename)
	}

	fmt.Println(url)

	http.Handle(fmt.Sprintf("/%v/%v", slug, filename), serveResponseHandler(response, filename, limits))

	return nil
}

func registerHandlers(args []string, slug string, limits *Limits) error {
	switch {
	case len(args) == 0 && !isFromPipe():
		err := errors.New("no file(s) specified and no data received from stdin")
		return err
	case len(args) == 0 && isFromPipe():
		registerHandler("", slug, limits)
	case len(args) != 0:
		for i := 0; i < len(args); i++ {
			_, err := os.Stat(args[i])
			if err != nil {
				return err
			}

			path, err := filepath.Abs(args[i])
			if err != nil {
				return err
			}

			registerHandler(path, slug, limits)
		}
	}

	return nil
}

func ServePage(args []string) {
	slug := generateRandomString(Length)

	limits := initializeLimits()

	err := registerHandlers(args, slug, limits)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-limits.channel

		os.Exit(0)
	}()

	err = http.ListenAndServe(":"+strconv.FormatInt(int64(Port), 10), nil)
	if err != nil {
		log.Fatal(err)
	}
}
