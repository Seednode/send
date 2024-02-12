## About

Sometimes, you just need to send someone a file, and don't want to stand up a whole webserver.

While this tool won't help you bypass NAT or anything like that, it does solve the initial issue.

Simply point this tool at one or more file(s), and it will generate URLs others can use to download them.

It also accepts input via stdin, optionally in combination with filenames.

Filenames are prefaced by a randomly generated slug, of configurable length. Set `-l|--length` to 0 to disable this.

If so inclined, you can also optionally obfuscate filenames with the `-r|--randomize` flag.

Static binary builds available [here](https://cdn.seedno.de/builds/send).

x86_64 and ARM Docker images of latest version: `oci.seedno.de/seednode/send:latest`.

Dockerfile available [here](https://git.seedno.de/seednode/send/raw/branch/master/docker/Dockerfile).

## Usage output
```
Generates a one-off download link for one or more specified file(s).

Usage:
  send [file]... [flags]

Flags:
  -b, --bind string                 address to bind to (default "0.0.0.0")
  -c, --count int                   number of times to serve the file(s)
  -d, --domain string               domain to use in returned urls
      --error-exit                  shut down webserver on error, instead of just printing error
  -h, --help                        help for send
  -l, --length int                  length of url slug and obfuscated filename(s) (default 6)
  -p, --port int                    port to listen on (default 8080)
  -r, --randomize                   randomize filename(s)
  -s, --scheme string               scheme to use in returned urls (default "http")
  -t, --timeout duration            shutdown after this length of time
      --timeout-interval duration   display remaining time in timeout every N seconds (default 1m0s)
  -u, --uri string                  full uri (overrides domain, scheme, and port)
  -v, --verbose                     log accessed files to stdout
      --version                     version for send
```
