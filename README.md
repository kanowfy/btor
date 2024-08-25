# Btor - command line BitTorrent client

Btor is a command line BitTorrent client that allows downloading using .torrent files
## Installation
### With `go install` (For Go 1.16 or later)
```shell
go install github.com/kanowfy/btor@latest
```

### Usage
Grab a torrent file from the Internet, let's say it's in `~/Downloads/example.torrent`. Then in the terminal:
```shell
# File will be saved to ~/example.txt
btor download ~/Downloads/example.torrent -o ~/example.txt
```
To view logs, either specify `-v` flag:
```shell
# Long form of the flag --verbose can also be used
btor download ~/Downloads/example.torrent -o ~/example.txt -v
```
Or check the logs in `$HOME/.local/share/btor/btor.log`:
```shell
cat $HOME/.local/share/btor/btor.log | less
```
For more commands and usage, use the -h flag:
```shell
# Print usage for all commands
btor -h

# Print usage for a specific command
btor peers -h
```

### Current limitations
- Only support HTTP trackers
- Only support single-file torrents

## License
[MIT LICENSE](LICENSE)
