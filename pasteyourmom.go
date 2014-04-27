package main

import (
    "flag"
    "io"
    "log"
    "math/rand"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/zenazn/goji"
    "github.com/zenazn/goji/web"
    "github.com/zenazn/goji/web/middleware"
)

const (
    idSet      string = "abcdefjhijklmnopqrstuvwxyzABCDEFJHIJKLMNOPQRSTUVWXYZ1234567890"
    idLength   int    = 5
    dataFolder string = "./"
)

var (
    stdoutLogger *log.Logger
    staticFiles  []string = []string{"index.html", "style.css"}
)

func genId() string {
    length := idLength
    endstr := make([]byte, length)
    i := 0
    for ; length > 0; length-- {
        endstr[i] = idSet[rand.Intn(len(idSet)-1)]
        i++
    }
    return string(endstr)
}

func isStaticFile(file string) bool {
    for _, element := range staticFiles {
        if file == element {
            return true
        }
    }
    return false
}

func root(c web.C, w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")

    file, err := os.Open(dataFolder + "index.html")
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        stdoutLogger.Print(err)
        return
    }

    _, err = io.Copy(w, file)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        return
    }
}

func getPaste(c web.C, w http.ResponseWriter, r *http.Request) {
    if isStaticFile(c.URLParams["id"]) {
        getStatic(c.URLParams["id"], w)
        return
    }
    file, err := os.Open(dataFolder + c.URLParams["id"] + ".paste")
    if os.IsNotExist(err) {
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        stdoutLogger.Print(err)
        return
    }

    _, err = io.Copy(w, file)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        return
    }
}

func createPaste(w http.ResponseWriter, r *http.Request) {
    id := genId()
    file, err := os.Create(dataFolder + id + ".paste")
    for os.IsExist(err) {
        id = genId()
        file, err = os.Create(dataFolder + id + ".paste")
    }
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        stdoutLogger.Print(err)
        return
    }

    text := r.FormValue("text")
    if text == "" {
        http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
        return
    }

    n, err := file.Write([]byte(text))
    if err != nil || n != len(text) {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, id, http.StatusSeeOther)
}

func getStatic(filename string, w http.ResponseWriter) {
    filenameSplitted := strings.Split(filename, ".")
    if len(filenameSplitted) > 1 {
        w.Header().Set("Content-Type", "text/"+filenameSplitted[len(filenameSplitted)-1])
    }
    file, err := os.Open(dataFolder + filename)
    if os.IsNotExist(err) {
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        stdoutLogger.Print(err)
        return
    }
    _, err = io.Copy(w, file)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        return
    }
}

func main() {
    xRealIP := flag.Bool("realip", false, "use X-Real-IP for getting the IP adress of the client")
    flag.Parse()
    stdoutLogger = log.New(os.Stderr, "", log.Flags())
    rand.Seed(time.Now().Unix())
    if *xRealIP {
        goji.Insert(middleware.RealIP, middleware.Logger)
    }
    goji.Get("/", root)
    goji.Get("/:id", getPaste)
    goji.Post("/", createPaste)
    goji.Serve()
}
