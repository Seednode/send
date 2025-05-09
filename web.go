/*
Copyright © 2025 Seednode <seednode@seedno.de>
*/

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
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
	ErrInvalidCount     = errors.New("count must be a non-negative integer")
	ErrInvalidLength    = errors.New("length must be a non-negative integer")
	ErrInvalidPort      = errors.New("listen port must be an integer between 1 and 65535 inclusive")
	ErrInvalidTimeout   = errors.New("timeout interval must be longer than timeout")
	ErrInvalidTLSConfig = errors.New("TLS certificate and keyfile must both be specified to enable HTTPS")
	ErrNoFile           = errors.New("no files specified and no data received from stdin")
)

const (
	logDate       = "2006-01-02T15:04:05.000-07:00"
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

func securityHeaders(w http.ResponseWriter) {
	w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-site")
	w.Header().Set("Permissions-Policy", "geolocation=(), midi=(), sync-xhr=(), microphone=(), camera=(), magnetometer=(), gyroscope=(), fullscreen=(), payment=()")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Xss-Protection", "1; mode=block")
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

	builder := strings.Builder{}
	builder.Grow(length)

	for i, cache, remain := length-1, rand.Int64(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int64(), letterIdxMax
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

func updateCounter(limits *Limits) string {
	atomic.AddUint32(limits.counter, 1)
	counter := atomic.LoadUint32(limits.counter)
	if counter >= uint32(Count) {
		defer func() {
			limits.channel <- true
		}()
	}

	remaining := Count - int(counter)

	return fmt.Sprintf(" (%d remaining)", remaining)
}

func readStdin() ([]byte, error) {
	var response []byte

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		response = append(response, scanner.Bytes()...)
		response = append(response, "\n"...)

		err := scanner.Err()
		if err != nil {
			return nil, err
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

func serveResponse(w http.ResponseWriter, r http.Request, response []byte, fullpath string, limits *Limits) error {
	remaining := ""

	if Count != 0 {
		remaining = updateCounter(limits)
	}

	fmt.Printf("%s | %s => %s%s\n", time.Now().Format(logDate), fullpath, realIP(&r, true), remaining)

	w.Header().Set("Content-Type", http.DetectContentType(response))

	w.Header().Set("Content-Length", strconv.Itoa(len(response)))

	securityHeaders(w)

	_, err := w.Write(response)
	if err != nil {
		return err
	}

	return nil
}

func serveResponseHandler(response []byte, fullpath string, limits *Limits, errorChannel chan<- Error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		err := serveResponse(w, *r, response, fullpath, limits)
		if err != nil {
			errorChannel <- Error{Message: err, Host: realIP(r, true)}
		}
	}
}

func registerHandler(mux *httprouter.Router, path, slug string, limits *Limits, errorChannel chan<- Error) (url, fullpath string) {
	var filename string

	switch {
	case Randomize || path == "":
		filename = "/" + generateRandomString(Length)
	default:
		filename = "/" + filepath.Base(path)
	}

	var response []byte
	var err error

	if path == "" {
		response, err = readStdin()
		if err != nil {
			errorChannel <- Error{Message: err}

			return "", ""
		}

		fullpath = "<data from stdin>"
	} else {
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

		response, err = readFile(path)
		if err != nil {
			errorChannel <- Error{Message: err}

			return "", ""
		}
	}

	mux.GET(fmt.Sprintf("%s%s", slug, filename), serveResponseHandler(response, fullpath, limits, errorChannel))

	switch {
	case URL != "":
		return fmt.Sprintf("%s%s%s", URL, slug, filename), fullpath
	default:
		return fmt.Sprintf("%s://%s:%d%s%s", Scheme, Bind, Port, slug, filename), fullpath
	}
}

func registerHandlers(mux *httprouter.Router, args []string, slug string, limits *Limits, errorChannel chan<- Error) (urls, paths []string) {
	if len(args) == 0 && !isFromPipe() {
		errorChannel <- Error{Message: ErrNoFile}

		return urls, paths
	}

	if isFromPipe() {
		url, path := registerHandler(mux, "", slug, limits, errorChannel)
		if url != "" {
			urls = append(urls, url)
			paths = append(paths, path)
		}
	}

	for i := range args {
		url, path := registerHandler(mux, args[i], slug, limits, errorChannel)
		if url != "" {
			urls = append(urls, url)
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
				fmt.Printf("%s | Error: %s (<= %s)\n",
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

	urls, paths := registerHandlers(mux, args, "/"+generateRandomString(Length), limits, errorChannel)
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
			fmt.Printf("%s | Shutdown in %s\n", time.Now().Format(logDate), Timeout)

			ticker := time.NewTicker(TimeoutInterval)

			go func() {
				for range ticker.C {
					remains := Timeout - time.Since(startTime).Round(time.Second)

					if remains > 0 {
						fmt.Printf("%s | Shutdown in %s\n", time.Now().Format(logDate), Timeout-time.Since(startTime).Round(time.Second))
					}
				}
			}()
		}
	}

	var err error

	if TLSKey != "" && TLSCert != "" {
		fmt.Printf("%s | Listening on https://%s/\n",
			time.Now().Format(logDate),
			srv.Addr)

		err = srv.ListenAndServeTLS(TLSCert, TLSKey)
	} else {
		fmt.Printf("%s | Listening on http://%s/\n",
			time.Now().Format(logDate),
			srv.Addr)

		err = srv.ListenAndServe()
	}

	fmt.Printf("%s | Shutting down...\n", time.Now().Format(logDate))

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
