/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/julienschmidt/httprouter"
)

var (
	ErrInvalidCount   = errors.New("count must be a non-negative integer")
	ErrInvalidLength  = errors.New("length must be a non-negative integer")
	ErrInvalidPort    = errors.New("listen port must be an integer between 1 and 65535 inclusive")
	ErrInvalidTimeout = errors.New("timeout interval must be longer than timeout")
	ErrNoFile         = errors.New("no file(s) specified and no data received from stdin")
)

const (
	logDate       = `2006-01-02T15:04:05.000-07:00`
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

type Limits struct {
	channel chan bool
	counter *uint32
}

type Error struct {
	Message error
	Host    string
	Fatal   bool
}

func isFromPipe() bool {
	f, _ := os.Stdin.Stat()

	if (f.Mode() & os.ModeCharDevice) == 0 {
		return true
	} else {
		return false
	}
}

func generateRandomString(length int) string {
	if length < 1 {
		return ""
	}

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

	return "/" + builder.String()
}

func updateCounter(limits *Limits) {
	atomic.AddUint32(limits.counter, 1)
	counter := atomic.LoadUint32(limits.counter)
	if counter >= uint32(Count) {
		defer func() {
			limits.channel <- true
		}()
	}

	remaining := Count - int(counter)

	if remaining != 0 {
		fmt.Printf("%s | %d copies remaining\n", time.Now().Format(logDate), remaining)
	} else {
		fmt.Printf("%s | All copies sent\n", time.Now().Format(logDate))
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

func realIP(r *http.Request, includePort bool) string {
	fields := strings.SplitAfter(r.RemoteAddr, ":")

	host := strings.TrimSuffix(strings.Join(fields[:len(fields)-1], ""), ":")
	port := fields[len(fields)-1]

	if host == "" {
		return r.RemoteAddr
	}

	cfIP := r.Header.Get("Cf-Connecting-Ip")
	xRealIP := r.Header.Get("X-Real-Ip")

	switch {
	case cfIP != "" && includePort:
		return cfIP + ":" + port
	case cfIP != "":
		return cfIP
	case xRealIP != "" && includePort:
		return xRealIP + ":" + port
	case xRealIP != "":
		return xRealIP
	case includePort:
		return host + ":" + port
	default:
		return host
	}
}

func serveResponse(w http.ResponseWriter, r http.Request, response []byte, filename, fullpath string, limits *Limits) error {
	fmt.Printf("%s | Serving %s to %s\n", time.Now().Format(logDate), fullpath, realIP(&r, true))

	if Count != 0 {
		updateCounter(limits)
	}

	w.Header().Set("Content-Type", http.DetectContentType(response))

	w.Header().Set("Content-Length", strconv.Itoa(len(response)))

	_, err := w.Write(response)
	if err != nil {
		return err
	}

	return nil
}

func serveResponseHandler(response []byte, filename, fullpath string, limits *Limits, errorChannel chan<- Error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		err := serveResponse(w, *r, response, filename, fullpath, limits)
		if err != nil {
			errorChannel <- Error{Message: err, Host: realIP(r, true)}
		}
	}
}

func registerHandler(mux *httprouter.Router, path, slug string, limits *Limits, errorChannel chan<- Error) (url, fullpath string) {
	var filename string

	f, err := os.Stat(path)
	if err != nil {
		errorChannel <- Error{Message: err}

		return "", ""
	}
	if f.IsDir() {
		return "", ""
	}

	fullpath, err = filepath.Abs(path)
	if err != nil {
		errorChannel <- Error{Message: err}

		return "", ""
	}

	switch {
	case Randomize || path == "":
		filename = generateRandomString(Length)
	default:
		filename = "/" + filepath.Base(path)
	}

	var response []byte

	if path == "" {
		response, err = readStdin()
		if err != nil {
			errorChannel <- Error{Message: err}

			return "", ""
		}
	} else {
		response, err = readFile(path)
		if err != nil {
			errorChannel <- Error{Message: err}

			return "", ""
		}
	}

	mux.GET(fmt.Sprintf("%s%s", slug, filename), serveResponseHandler(response, filename, fullpath, limits, errorChannel))

	switch {
	case URL != "":
		return fmt.Sprintf("%s%s%s", URL, slug, filename), fullpath
	default:
		return fmt.Sprintf("http://%s:%d%s%s", Bind, Port, slug, filename), fullpath
	}
}

func registerHandlers(mux *httprouter.Router, args []string, slug string, limits *Limits, errorChannel chan<- Error) (urls, paths []string) {
	if len(args) == 0 && !isFromPipe() {
		errorChannel <- Error{Message: ErrNoFile}

		return urls, paths
	}

	for i := range args {
		url, path := registerHandler(mux, args[i], slug, limits, errorChannel)
		if url != "" {
			urls = append(urls, url)
		}
		if path != "" {
			paths = append(paths, path)
		}
	}

	if isFromPipe() {
		url, path := registerHandler(mux, "", slug, limits, errorChannel)
		if url != "" {
			urls = append(urls, url)
		}
		if path != "" {
			paths = append(paths, path)
		}
	}

	return urls, paths
}

func ServePage(args []string) error {
	startTime := time.Now()

	mux := httprouter.New()

	srv := &http.Server{
		Addr:         net.JoinHostPort(Bind, strconv.Itoa(Port)),
		Handler:      mux,
		IdleTimeout:  10 * time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Minute,
	}

	errorChannel := make(chan Error)

	go func() {
		for err := range errorChannel {
			if err.Host == "" {
				fmt.Printf("%s | Error: %s\n",
					time.Now().Format(logDate),
					err.Message)
			} else {
				fmt.Printf("%s | Error: %s (from %s)\n",
					time.Now().Format(logDate),
					err.Message,
					err.Host)
			}

			if ErrorExit || err.Fatal {
				srv.Shutdown(context.Background())

				break
			}
		}
	}()

	limits := &Limits{
		channel: make(chan bool, 1),
		counter: new(uint32),
	}

	go func() {
		<-limits.channel

		err := srv.Shutdown(context.Background())

		errorChannel <- Error{Message: err}
	}()

	if Profile {
		registerProfileHandlers(mux)
	}

	urls, paths := registerHandlers(mux, args, generateRandomString(Length), limits, errorChannel)
	if len(urls) == 0 || len(paths) == 0 {
		errorChannel <- Error{Message: ErrNoFile, Fatal: true}
	}

	for i := range urls {
		fmt.Printf("%s | %s -> %s\n",
			time.Now().Format(logDate),
			urls[i],
			paths[i])
	}

	if Timeout != 0 {
		time.AfterFunc(Timeout, func() {
			err := srv.Shutdown(context.Background())

			errorChannel <- Error{Message: err}
		})

		if TimeoutInterval > 0 {
			fmt.Printf("%s | Shutting down in %s\n", time.Now().Format(logDate), Timeout)

			ticker := time.NewTicker(TimeoutInterval)

			go func() {
				for range ticker.C {
					left := Timeout - time.Since(startTime).Round(time.Second)

					if left > 0 {
						fmt.Printf("%s | Shutting down in %s\n", time.Now().Format(logDate), Timeout-time.Since(startTime).Round(time.Second))
					}
				}
			}()
		}
	}

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	fmt.Printf("%s | Shutting down...\n", time.Now().Format(logDate))

	return nil
}
