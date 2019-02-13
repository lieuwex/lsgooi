package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
)

const (
	port     = "8080"
	dir      = "/files"
	authfile = "/auth"
	urlfmt   = "https://f.lieuwe.xyz/%s/%s"
)

var User string
var Pass string

// Item represents an gooid item on disk.
type Item struct {
	ID   string
	Name string
	Size string
	Date time.Time
	URL  string
}

var itemMap map[string]Item

func readItems(dir string, refresh bool) (map[string]Item, error) {
	m := itemMap
	if refresh {
		m = make(map[string]Item)
	}

	items, err := ioutil.ReadDir(dir)
	if err != nil {
		return m, err
	}

	for _, f := range items {
		if strings.HasSuffix(f.Name(), "-fname") {
			continue
		}

		id := f.Name()

		if _, has := m[id]; has {
			continue
		}

		p := path.Join(dir, id+"-fname")
		fname, err := ioutil.ReadFile(p)
		if err != nil {
			return m, err
		}

		m[id] = Item{
			ID:   id,
			Name: strings.TrimSpace(string(fname)),
			Size: humanize.Bytes(uint64(f.Size())),
			Date: f.ModTime(),
			URL:  fmt.Sprintf(urlfmt, id, fname),
		}
	}

	return m, nil
}

func check(user, pass string) bool {
	return User == user && Pass == pass
}

func root(w http.ResponseWriter, r *http.Request) {
	if user, pass, ok := r.BasicAuth(); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("auth required\n"))
		return
	} else if !check(user, pass) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("incorrect user/pass\n"))
		return
	}

	items := make([]Item, 0, len(itemMap))
	for _, v := range itemMap {
		items = append(items, v)
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Date.Before(items[j].Date)
	})

	const tpl = `
<!doctype html>
<html>
	<head>
		<meta charset="utf-8">
		<title>gooid</title>
		<style>
			body {
				font-family: monospace;
			}

			.file:not(:last-child) {
				border-bottom: 1px black solid;
			}
		</style>
	</head>
	<body>
		{{range $v := .}}
			<a href="{{$v.URL}}">
				<div class="file">
					<div class="fname">{{$v.Name}}</div>
					<div class="size">{{$v.Size}}</div>
					<div class="date">{{$v.Date}}</div>
				</div>
			</a>
		{{end}}
	</body>
</html>`

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		panic(err)
	}

	if err := t.Execute(w, items); err != nil {
		panic(err)
	}
}

func main() {
	var err error
	itemMap, err = readItems(dir, true)
	if err != nil {
		panic(err)
	}

	auth, err := ioutil.ReadFile(authfile)
	if err != nil {
		panic(err)
	}
	for i, line := range strings.Split(string(auth), "\n") {
		line = strings.TrimSpace(line)
		switch i {
		case 0:
			User = line
		case 1:
			Pass = line
		}
	}

	log.Printf("read %d file(s)", len(itemMap))

	http.HandleFunc("/", root)

	log.Printf("listening on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
