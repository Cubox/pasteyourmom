package main

import (
    "encoding/json"
    "flag"
    "io"
    "io/ioutil"
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

type configuration struct {
    DataFolder  string
    FormTitle   string
    IdLength    int
    IdSet       string
    StaticFiles []string
    RealIP      bool
}

var (
    conf configuration = configuration{
        "./",
        "content",
        5,
        "abcdefjhijklmnopqrstuvwxyzABCDEFJHIJKLMNOPQRSTUVWXYZ1234567890",
        []string{"index.html", "style.css"},
        false}
    stdoutLogger *log.Logger
)

func dumpDefaultConf() {
    buf, err := json.Marshal(conf)
    if err != nil {
        stdoutLogger.Fatal(err)
    }

    os.Stdout.Write(buf)
    os.Stdout.Write([]byte("\n"))
}

func setConf(confFile string) {
    content, err := ioutil.ReadFile(confFile)
    if err != nil {
        stdoutLogger.Fatal(err)
    }

    err = json.Unmarshal(content, &conf)
    if err != nil {
        stdoutLogger.Fatal(err)
    }
}

func genId() string {
    length := conf.IdLength
    endstr := make([]byte, length)

    i := 0

    for ; length > 0; length-- {
        endstr[i] = conf.IdSet[rand.Intn(len(conf.IdSet)-1)]
        i++
    }

    return string(endstr)
}

func isStaticFile(file string) bool {
    for _, element := range conf.StaticFiles {
        if file == element {
            return true
        }
    }

    return false
}

func root(c web.C, w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")

    file, err := os.Open(conf.DataFolder + "index.html")
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        stdoutLogger.Print(err)
        return
    }

    _, err = io.Copy(w, file)
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        stdoutLogger.Print(err)
        return
    }
}

func getPaste(c web.C, w http.ResponseWriter, r *http.Request) {
    if isStaticFile(c.URLParams["id"]) {
        getStatic(c.URLParams["id"], w)
        return
    }

    file, err := os.Open(conf.DataFolder + c.URLParams["id"] + ".paste")
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
        stdoutLogger.Print(err)
        return
    }
}

func createPaste(w http.ResponseWriter, r *http.Request) {
    id := genId()

    file, err := os.Create(conf.DataFolder + id + ".paste")
    for os.IsExist(err) {
        id = genId()
        file, err = os.Create(conf.DataFolder + id + ".paste")
    }
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        stdoutLogger.Print(err)
        return
    }

    text := r.FormValue(conf.FormTitle)
    if text == "" {
        http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
        return
    }

    n, err := file.Write([]byte(text))
    if err != nil || n != len(text) {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        stdoutLogger.Print(err)
        return
    }

    http.Redirect(w, r, id, http.StatusSeeOther)

    n, err = w.Write([]byte(r.Host + "/" + id + "\n"))
    if err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        stdoutLogger.Print(err)
        return
    }
}

func getStatic(filename string, w http.ResponseWriter) {
    filenameSplitted := strings.Split(filename, ".")
    if len(filenameSplitted) > 1 {
        w.Header().Set("Content-Type", "text/"+filenameSplitted[len(filenameSplitted)-1])
    }

    file, err := os.Open(conf.DataFolder + filename)
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
        stdoutLogger.Print(err)
        return
    }
}

func main() {
    var dumpConf bool
    var confFile string

    flag.BoolVar(&dumpConf, "dumpconf", false, "dump the default configuration file")
    flag.StringVar(&confFile, "conf", "", "use this alternative configuration file")
    flag.Parse()

    stdoutLogger = log.New(os.Stderr, "", log.Flags())

    if dumpConf {
        dumpDefaultConf()
        return
    }

    rand.Seed(time.Now().Unix())

    if confFile != "" {
        setConf(confFile)
    }

    if conf.RealIP {
        goji.Insert(middleware.RealIP, middleware.Logger)
    }

    goji.Get("/", root)
    goji.Get("/:id", getPaste)
    goji.Post("/", createPaste)
    goji.Serve()
}
