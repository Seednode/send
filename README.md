## About

Sometimes, you just need to send someone a file, and don't want to stand up a whole webserver.

While this tool won't help you bypass NAT or anything like that, it does solve the initial issue.

Simply point this tool at one or more file, and it will generate a URL others can use to download it.

If no files are specified, it will read from stdin.

Filenames are prefaced by a randomly generated slug, of configurable length.

If so inclined, you can also optionally obfuscate filenames.

Builds available [here](https://cdn.seedno.de/builds/send).

## Usage output
```
Usage:
  send [file] [file]... [flags]
  send [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print version

Flags:
  -c, --count uint32    number of times to serve the file(s)
  -d, --domain string   domain to use in returned urls (default "localhost")
  -h, --help            help for send
  -l, --length uint16   length of url slug and obfuscated filename(s) (default 6)
  -p, --port uint16     port to listen on (default 8080)
  -r, --randomize       randomize filename(s)
  -s, --scheme string   scheme to use in returned urls (default "http")
  -v, --verbose         log accessed files to stdout

Use "send [command] --help" for more information about a command.
```
