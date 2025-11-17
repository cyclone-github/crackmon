[![Readme Card](https://github-readme-stats.vercel.app/api/pin/?username=cyclone-github&repo=crackmon&theme=gruvbox)](https://github.com/cyclone-github/crackmon/)

[![Go Report Card](https://goreportcard.com/badge/github.com/cyclone-github/crackmon)](https://goreportcard.com/report/github.com/cyclone-github/crackmon)
[![GitHub issues](https://img.shields.io/github/issues/cyclone-github/crackmon.svg)](https://github.com/cyclone-github/crackmon/issues)
[![License](https://img.shields.io/github/license/cyclone-github/crackmon.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/cyclone-github/crackmon.svg)](https://github.com/cyclone-github/crackmon/releases)
<!-- [![Go Reference](https://pkg.go.dev/badge/github.com/cyclone-github/crackmon.svg)](https://pkg.go.dev/github.com/cyclone-github/crackmon) -->

# crackmon
Hashcat & mdxfind wrapper tool to stop current attack if crack rate drops below threshold.

### Usage:
Default: -time 1m -crack 1
```
./crackmon ./hashcat {hashcat args}
./crackmon ./mdxfind {mdxfind args}
```
Custom: -time 5m -crack 100
```
./crackmon -t 5 -c 100 ./hashcat {hashcat args}
./crackmon -t 5 -c 100 ./mdxfind {mdxfind args}
```
For more info:
```
./crackmon -help
Examples:

Defaults to -time 1m -crack 1
./crackmon ./hashcat {hashcat args}
./crackmon ./mdxfind {mdxfind args}

Custom: -time 5m -crack 100
./crackmon -t 5 -c 100 ./hashcat {hashcat args}
./crackmon -t 5 -c 100 ./mdxfind {mdxfind args}

All flags:
	-t         minimum runtime in minutes
	-c         cumulative average cracks threshold
	-debug     enable debug output
	-help      show this help menu
	-version   show version info

Supported hashcat attacks:
	-a 0       straight
	-a 1       combination
	-a 9       associated

Partially supported hashcat attacks:
	-a 3       mask
	-a 6       hybrid
	-a 7       hybrid

Unsupported hashcat flags:
	-i         incremental
	--status-json
```

For troubleshooting, run with -debug flag
```
./crackmon -debug ./hashcat...
```
### Notes:
- Developed and tested on debian 12/13 and Windows 11 Terminal
- Designed for running hashcat attacks `-a 0, 1, 9`. 
- Supports `-a 3, 6, 7`, but does not currently support hashcat mask files or `-incremental` due to how hashcat handles sessions when running -i or mask files.
- Crackmon v0.3.0 added beta support for mdxfind.
### Changelog:
https://github.com/cyclone-github/crackmon/blob/main/CHANGELOG.md

### Install latest release:
- `go install github.com/cyclone-github/crackmon@latest`
### Install from latest source code (bleeding edge):
- `go install github.com/cyclone-github/crackmon@main`