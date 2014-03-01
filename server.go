package main

import (
    "github.com/codegangsta/martini"
    "github.com/martini-contrib/cors"
    "github.com/codegangsta/martini-contrib/render"
    "github.com/wtolson/go-taglib"
    "encoding/json"
    "path/filepath"
    "net/http"
    "regexp"
    "os"
)

func main() {
    m := martini.Classic()
    m.Use(martini.Static("."))
    m.Use(cors.Allow(&cors.Options{
        AllowOrigins:     []string{"http://*"},
        AllowMethods:     []string{"GET"},
        AllowHeaders:     []string{"Origin"},
    }))
    m.Use(render.Renderer())

    m.Get("/", func(r render.Render) {
        r.JSON(200, songs)
    })

    m.Get("/search", func(w http.ResponseWriter, r *http.Request) {
        search := r.FormValue("search")
        matches := QuerySongs(search)
        enc := json.NewEncoder(w)
        enc.Encode(matches)
    })

    m.Run()
}

func QuerySongs(search string) []Song {
    matches := make([]Song, 0)
    if search == "" {
        return matches
    }
    for _, song := range songs {
        s := []string{song.Name, song.Artist, song.Album, song.Genre}
        for _, attr := range s {
            if match, _ := regexp.MatchString(search, attr); match {
                matches = append(matches, song)
                break
            }
        }
    }
    return matches
}

func ClosureHackage(songs *[]Song) func(s string, f os.FileInfo, err error) error {
    return func(path string, f os.FileInfo, err error) error {
        re := regexp.MustCompile(`\.(mp3|m4a)$`)
        if match := re.FindString(path); match != "" {
            f, _ := taglib.Read(path)
            defer f.Close()
            song := Song{}
            song.Name = f.Title()
            song.Artist = f.Artist()
            song.Album = f.Album()
            song.Year = f.Year()
            song.Track = f.Track()
            song.Genre = f.Genre()
            song.Length = int(f.Length().Seconds())
            song.Path = "/" + path
            *songs = append(*songs, song)
        }
        return nil
    }
}

var songs = Songs()
func Songs() []Song {
    songs := make([]Song, 0)
    FindMp3s := ClosureHackage(&songs)
    filepath.Walk(".", FindMp3s)
    return songs
}

type Song struct {
    Name string
    Artist string
    Album string
    Year int
    Track int
    Genre string
    Length int
    Path string
}
