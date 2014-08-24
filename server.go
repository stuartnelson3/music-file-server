package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gorilla/handlers"
	"github.com/gorilla/pat"
	"github.com/wtolson/go-taglib"
)

var (
	mux   = pat.New()
	port  = flag.String("p", "3000", "address to bind the server on")
	songs []Song
)

func main() {
	flag.Parse()

	songs = allSongs()

	mux.Get("/search", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		matches := querySongs(r.FormValue("search"))
		json.NewEncoder(w).Encode(matches)
	}))

	mux.Get("/", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	}))

	handler := handlers.LoggingHandler(os.Stdout, mux)
	log.Printf("listening on %s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, handler))
}

func corsHandler(fn func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		fn(w, r)
	}
}

func querySongs(search string) []Song {
	if search == "" {
		matches := make([]Song, 0)
		return matches
	}
	return matchesInSongSlice(songs, search)
}

func matchesInSongSlice(songs []Song, search string) []Song {
	matches := make([]Song, 0)
	for _, song := range songs {
		s := []string{song.Name, song.Artist, song.Album, song.Genre}
		for _, attr := range s {
			if match, _ := regexp.MatchString("(?i)"+search, attr); match {
				matches = append(matches, song)
				break
			}
		}
	}
	return matches
}

func closure(songs *[]Song) func(s string, f os.FileInfo, err error) error {
	return func(path string, f os.FileInfo, err error) error {
		re := regexp.MustCompile(`\.(mp3|m4a)$`)
		if match := re.FindString(path); match != "" {
			f, _ := taglib.Read(path)
			defer f.Close()
			ip, _ := localIP()
			song := Song{
				Name:   f.Title(),
				Artist: f.Artist(),
				Album:  f.Album(),
				Year:   f.Year(),
				Track:  f.Track(),
				Genre:  f.Genre(),
				Length: int(f.Length().Seconds()),
				Url:    "http://" + ip.String() + ":" + *port + "/" + path,
			}
			*songs = append(*songs, song)
		}
		return nil
	}
}

func localIP() (net.IP, error) {
	tt, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, t := range tt {
		aa, err := t.Addrs()
		if err != nil {
			return nil, err
		}
		for _, a := range aa {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			v4 := ipnet.IP.To4()
			if v4 == nil || v4[0] == 127 { // loopback address
				continue
			}
			return v4, nil
		}
	}
	return nil, errors.New("cannot find local IP address")
}

func allSongs() []Song {
	songs := make([]Song, 0)
	FindMp3s := closure(&songs)
	filepath.Walk(".", FindMp3s)
	return songs
}

type Song struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Year   int    `json:"year"`
	Track  int    `json:"track"`
	Genre  string `json:"genre"`
	Length int    `json:"length"`
	Url    string `json:"url"`
}
