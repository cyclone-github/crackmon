# crackmon
Hashcat wrapper tool to bypass current attack if crack rate drops below threshold.

Inspiration by: https://github.com/justpretending/avgdrop

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
        -t              time threshold in minutes
        -c              average cracks threshold per -t
        -debug          enable debug output
        -help           show this help menu
        -version        show version info
```
For troubleshooting, run with -debug flag
```
./crackmon -debug ./hashcat...
```
### Version:
- v2023-10-01.1700; initial github release
- v2023-10-02.1030-debug; fixed timeThreshold; added more debugging info
- v2023-10-04.1545-winpty; added pty support for windows; debug flag; changed CUR to AVG

### Notes:
- Compiled and tested on debian 12 and Windows 11 Terminal

### Compile from source code info:
- https://github.com/cyclone-github/scripts/blob/main/intro_to_go.txt
