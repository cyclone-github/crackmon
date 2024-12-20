[![Readme Card](https://github-readme-stats.vercel.app/api/pin/?username=cyclone-github&repo=crackmon&theme=gruvbox)](https://github.com/cyclone-github/crackmon/)

[![Go Report Card](https://goreportcard.com/badge/github.com/cyclone-github/crackmon)](https://goreportcard.com/report/github.com/cyclone-github/crackmon)
[![GitHub issues](https://img.shields.io/github/issues/cyclone-github/crackmon.svg)](https://github.com/cyclone-github/crackmon/issues)
[![License](https://img.shields.io/github/license/cyclone-github/crackmon.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/cyclone-github/crackmon.svg)](https://github.com/cyclone-github/crackmon/releases)
<!-- [![Go Reference](https://pkg.go.dev/badge/github.com/cyclone-github/crackmon.svg)](https://pkg.go.dev/github.com/cyclone-github/crackmon) -->

# crackmon
Hashcat wrapper tool to bypass current attack if crack rate drops below threshold.

Inspiration by: https://launchpad.net/avgdrop

### Usage:
Default: -time 1m -crack 1
```
./crackmon ./hashcat {hashcat args}
```
Custom: -time 2m -crack 100
```
./crackmon -t 2 -c 100 ./hashcat {hashcat args}
```
For more info:
```
./crackmon -help
Examples:

(Defaults to -time 1m -crack 1)
./crackmon ./hashcat {hashcat args}

Custom: -time 2m -crack 100
./crackmon -t 2 -c 100 ./hashcat {hashcat args}

All flags:
	-t      	minimum runtime in minutes
	-c      	cumulative average cracks threshold
	-debug  	enable debug output
	-help   	show this help menu
	-version	show version info
```
For troubleshooting, run with -debug flag
```
./crackmon -debug ./hashcat...
```
### Version:
- v2023-10-01.1700; initial github release
- v2023-10-02.1030-debug; fixed timeThreshold; added more debugging info
- v2023-10-04.1545-winpty; added pty support for windows; debug flag; changed CUR to AVG
- v2023-10-07.1520-pty; refactored code, added logic for hashcat Paused, Running and Stopped status, added support for user sending commands to hashcat
- v2023-10-13.1445; fixed https://github.com/cyclone-github/crackmon/issues/4; refactored sendX commands

### Notes:
- Compiled and tested on debian 12 and Windows 11 Terminal

### Compile from source:
- If you want the latest features, compiling from source is the best option since the release version may run several revisions behind the source code.
- This assumes you have Go and Git installed
  - `git clone https://github.com/cyclone-github/crackmon.git`
  - `cd crackmon`
  - `go mod init crackmon`
  - `go mod tidy`
  - `go build -ldflags="-s -w" .`
- Compile from source code how-to:
  - https://github.com/cyclone-github/scripts/blob/main/intro_to_go.txt
