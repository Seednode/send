/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
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

const LOGDATE string = "2006-01-02T15:04:05.000000000-07:00"
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

type Limits struct {
	channel chan bool
	counter *uint32
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

func serveFile(w http.ResponseWriter, r http.Request, path string, limits *Limits) error {
	atomic.AddUint32(limits.counter, 1)
	counter := atomic.LoadUint32(limits.counter)
	if counter >= Count && Count != 0 {
		defer func() {
			limits.channel <- true
		}()
	}

	var startTime time.Time
	if Verbose {
		startTime = time.Now()
		fmt.Printf("%v | %v requested %v", startTime.Format(LOGDATE), r.RemoteAddr, path)
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	w.Write(buf)

	if Verbose {
		fmt.Printf(" (Finished in %v)\n", time.Now().Sub(startTime).Round(time.Microsecond))
	}

	return nil
}

func serveFileHandler(path string, limits *Limits) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := serveFile(w, *r, path, limits)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func doNothing(http.ResponseWriter, *http.Request) {}

func ServePage(args []string) {
	path, err := filepath.Abs(args[0])
	if err != nil {
		log.Fatal(err)
	}

	_, err = os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	limits := initializeLimits()

	slug := generateRandomString(Length)

	var filename string
	switch {
	case Randomize:
		filename = generateRandomString(Length)
	default:
		filename = filepath.Base(path)
	}

	var url string
	switch URI {
	case "":
		url = fmt.Sprintf("%v://%v:%v/%v/%v", Scheme, Domain, Port, slug, filename)
	default:
		url = fmt.Sprintf("%v/%v/%v", URI, slug, filename)
	}
	fmt.Println(url)

	http.Handle(fmt.Sprintf("/%v/%v", slug, filename), serveFileHandler(path, limits))
	http.HandleFunc("/favicon.ico", doNothing)

	go func() {
		<-limits.channel

		os.Exit(0)
	}()

	err = http.ListenAndServe(":"+strconv.FormatInt(int64(Port), 10), nil)
	if err != nil {
		log.Fatal(err)
	}
}
