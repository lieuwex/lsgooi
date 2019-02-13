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
	urlfmt   = "https://f.lieuwe.xyz/vang/%s/%s"
)

var User string
var Pass string

// Item represents an gooid item on disk.
type Item struct {
	ID   string
	Name string
	Size uint64
	Date time.Time
	URL  string
}

func (item Item) SizeString() string {
	return humanize.Bytes(item.Size)
}

func (item Item) DateString() string {
	return item.Date.Format("2006-01-02 15:04:05")
}

var itemMap map[string]Item

func readItems(dir string, m map[string]Item) (map[string]Item, error) {
	if m == nil {
		m = make(map[string]Item)
	}

	items, err := ioutil.ReadDir(dir)
	if err != nil {
		return m, err
	}

	for _, f := range items {
		if strings.HasSuffix(f.Name(), "-fname") || f.Name() == "startid" {
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
			Size: uint64(f.Size()),
			Date: f.ModTime(),
			URL:  fmt.Sprintf(urlfmt, id, fname),
		}
	}

	return m, nil
}

func root(w http.ResponseWriter, r *http.Request) {
	if user, pass, ok := r.BasicAuth(); !ok {
		w.Header().Add("WWW-Authenticate", "Basic")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("auth required\n"))
		return
	} else if User != user || Pass != pass {
		w.Header().Add("WWW-Authenticate", "Basic")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("incorrect user/pass\n"))
		return
	}

	items := make([]Item, 0, len(itemMap))
	for _, v := range itemMap {
		items = append(items, v)
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[j].Date.Before(items[i].Date)
	})

	const tpl = `
<!doctype html>
<html>
	<head>
		<meta charset="utf-8">
		<title>lsgooi</title>
		<style>
			body {
				font-family: monospace;
			}

			a {
				color: black;
				text-decoration: none;
			}

			.file {
				border-bottom: 1px lightgray solid;
				margin: 10px;
				padding-bottom: 10px;
				display: flex;
			}

			.fname {
				width: 300px;
				margin-right: 10px;
			}

			.size {
				width: 100px;
			}
		</style>
	</head>
	<body>
		{{range $v := .}}
			<a href="{{$v.URL}}">
				<div class="file">
					<div class="fname">{{$v.Name}}</div>
					<div class="size">{{$v.SizeString}}</div>
					<div class="date">{{$v.DateString}}</div>
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
	go func() {
		prev := 0
		for {
			var err error

			itemMap, err = readItems(dir, itemMap)
			if err != nil {
				panic(err)
			}
			if l := len(itemMap); l != prev {
				log.Printf("read %d file(s)", l)
				prev = l
			}

			time.Sleep(10 * time.Second)
		}
	}()

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

	http.HandleFunc("/", root)

	log.Printf("listening on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
