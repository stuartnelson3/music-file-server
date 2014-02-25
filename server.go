package main

import (
    "github.com/codegangsta/martini"
    "github.com/martini-contrib/cors"
    "github.com/codegangsta/martini-contrib/render"
    "github.com/ascherkus/go-id3/src/id3"
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
            metadata := id3.File{}
            if match == ".mp3" {
                mp3File, err := os.Open(path)
                if err != nil {
                    return err
                }
                metadata = *id3.Read(mp3File)
            } else {
                metadata.Name = path
            }
            song := Song{Metadata: metadata, Path: "/" + path}
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
    Metadata id3.File
    Path string
}
