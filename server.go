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
    "errors"
    "net"
    "os"
)

func main() {
    m := martini.Classic()
    m.Use(martini.Static("."))
    m.Use(cors.Allow(&cors.Options{
        AllowOrigins:     []string{"http://*", "https://*"},
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
    splitSongs := splitSlice(songs)
    results := make(chan []Song, len(splitSongs))
    defer close(results)
    for i := 0; i < len(splitSongs); i++ {
        go matchesInSongSlice(splitSongs[i], search, results)
    }
    // should this be used for the timeout?
    // for {
    //     var done bool
    //     select {
    //     case songs := <-results:
    //         matches = append(matches, songs...)
    //     case <-time.After(time.Second):
    //         done = true
    //     }
    //     if done { break }
    // }
    for i := 0; i < len(splitSongs); i++ {
        songs := <-results
        matches = append(matches, songs...)
    }
    return matches
}

func matchesInSongSlice(songs []Song, search string, results chan []Song) {
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
    results<- matches
}

func splitSlice(songs []Song) [][]Song {
    n := 1
    l := len(songs)
    for i := 2; i < 13; i++ {
        if l % i == 0 {
            n = i
        }
    }
    splitSlices := make([][]Song, 0)
    for i := 0; i < n; i++ {
        chunk := len(songs)/n
        start := i*chunk
        finish := (i+1)*chunk
        splitSlices = append(splitSlices, songs[start:finish])
    }
    return splitSlices
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
            ip, _ := localIP()
            port := "3000"
            if envPort := os.Getenv("PORT"); envPort != "" {
                port = envPort
            }
            song.Url = "http://" + ip.String() + ":" + port + "/" + path
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

var songs = Songs()
func Songs() []Song {
    songs := make([]Song, 0)
    FindMp3s := ClosureHackage(&songs)
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
