mp3 file server

## usage
clone the repo and then build the binary
```
$ go build
```
go to the folder that has mp3s and run the binary. the music server will walk
through the root and all child directories looking for mp3s and serve them up,
so be careful of where you run the binary.
