# crackmon
Hashcat wrapper tool to bypass current attack if crack rate drops below threshold

### Usage:
Default: -time 1m -crack 1
```
./crackmon ./hashcat {hashcat args}
```
Custom: -time 10m -crack 2
```
./crackmon -t 10 -c 2 ./hashcat {hashcat args}
```
### Note:
- Currently only supports Linux and Mac. Windows is on todo list, but is not currently working due to crappy PTY support on Windows terminal / Power Shell.

### Version:
- v2023-10-01.1700; initial github release

### Compile from source code info:
- https://github.com/cyclone-github/scripts/blob/main/intro_to_go.txt