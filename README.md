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

### Configuration
The following configuration methods are accepted, in order of highest to lowest priority:
- Command-line flags
- Environment variables

### Environment variables
Almost all options configurable via flags can also be configured via environment variables. 

The associated environment variable is the prefix `SEND_` plus the flag name, with the following changes:
- Leading hyphens removed
- Converted to upper-case
- All internal hyphens converted to underscores

For example:
- `--bind 127.0.0.1` becomes `SEND_BIND=127.0.0.1`
- `--interval 10s` becomes `SEND_INTERVAL=10s`
- `--randomize` becomes `SEND_RANDOMIZE=true`

## Usage output
```
Generates a one-off download link for one or more specified files.

Usage:
  send [file]... [flags]

Flags:
  -b, --bind string         address to bind to (default "0.0.0.0")
  -c, --count int           number of times to serve files
  -e, --exit                shut down webserver on error, instead of just printing error
  -h, --help                help for send
  -i, --interval duration   display remaining time in timeout at this interval (default 1m0s)
  -l, --length int          length of url slug and obfuscated filenames (default 6)
  -p, --port int            port to listen on (default 8080)
      --profile             register net/http/pprof handlers
  -r, --randomize           randomize filenames
  -s, --scheme string       scheme to use in returned URLs (default "http")
  -t, --timeout duration    shutdown after this length of time
      --tls-cert string     path to TLS certificate
      --tls-key string      path to TLS keyfile
  -u, --url string          use this value instead of <scheme>://<bind>:<port> in returned URLs
  -v, --version             version for send
```

## Building the Docker image
From inside the cloned repository, build the image using the following command:

`REGISTRY=<registry url> LATEST=yes TAG=alpine ./build-docker.sh`
