package main

import (
    "github.com/codegangsta/martini"
    "github.com/martini-contrib/cors"
    "github.com/codegangsta/martini-contrib/render"
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
    m.Use(render.Renderer(render.Options{
        Layout:     "layout",
        Delims: render.Delims{"{[{", "}]}"},
        Extensions: []string{".html"}}))

    m.Get("/", func(r render.Render) {
        songs := Songs()
        r.JSON(200, songs)
    })

    m.Run()
}

func ClosureHackage(payload *Payload) func(s string, f os.FileInfo, err error) error {
    return func(path string, f os.FileInfo, err error) error {
        re := regexp.MustCompile(`\.mp3$`)
        if match := re.FindString(path); match != "" {
            payload.Songs = append(payload.Songs, path)
        }
        return nil
    }
}

func Songs() Payload {
    songs := make([]string, 0)
    payload := Payload{Songs: songs}
    FindMp3s := ClosureHackage(&payload)
    filepath.Walk(".", FindMp3s)
    return payload
}

type Payload struct {
    Songs []string
}
