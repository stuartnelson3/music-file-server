package main

import (
    "github.com/codegangsta/martini"
    "github.com/martini-contrib/cors"
    "github.com/codegangsta/martini-contrib/render"
    "github.com/wtolson/go-taglib"
    "path/filepath"
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
        // consider moving songs outside of each get request so that the server
        // caches the songs found initially
        songs := Songs()
        r.JSON(200, songs)
    })

    m.Run()
}

func ClosureHackage(payload *Payload) func(s string, f os.FileInfo, err error) error {
    return func(path string, f os.FileInfo, err error) error {
        re := regexp.MustCompile(`\.(mp3|m4a)$`)
        if match := re.FindString(path); match != "" {
            f, _ := taglib.Read(path)
            song := Song{}
            song.Name = f.Title()
            song.Artist = f.Artist()
            song.Album = f.Album()
            song.Year = f.Year()
            song.Track = f.Track()
            song.Genre = f.Genre()
            song.Length = int(f.Length().Seconds())
            song.Path = "/" + path
            payload.Songs = append(payload.Songs, song)
        }
        return nil
    }
}

func Songs() Payload {
    songs := make([]Song, 0)
    payload := Payload{Songs: songs}
    FindMp3s := ClosureHackage(&payload)
    filepath.Walk(".", FindMp3s)
    return payload
}

type Payload struct {
    Songs []Song
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
