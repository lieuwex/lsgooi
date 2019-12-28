package main

import (
	"bytes"
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

// Item represents an gooid item on disk.
type Item struct {
	ID   string
	Name string
	Size uint64
	Date time.Time
	URL  string
}

// SizeString returns the size of the current item in a human friendly format.
func (item Item) SizeString() string {
	return humanize.Bytes(item.Size)
}

// DateString returns the modification date of the current item in a human and
// machine friendly format.
func (item Item) DateString() string {
	return item.Date.Format("2006-01-02 15:04:05")
}

var (
	// tpl contains the current compiled version of the page.
	tpl            []byte
	lastUpdateTime = time.Unix(0, 0)
	prevItemCount  = 0

	// correctUser is the correct username of the user.
	correctUser string
	// correctPass is the correct password of the user.
	correctPass string
)

// readItems reads the given gooi files directory for items, if prev is non-nil
// it will be used as a cache for existing files. If a file is removed from the
// directory it isn't included in the result, even if prev does contain it.
func readItems(dir string, prev map[string]Item) (map[string]Item, error) {
	items, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	m := make(map[string]Item)
	for _, f := range items {
		if strings.HasSuffix(f.Name(), "-fname") || f.Name() == "startid" {
			continue
		}

		id := f.Name()

		if val, has := prev[id]; has {
			m[id] = val
			continue
		}

		fname, err := ioutil.ReadFile(path.Join(dir, id+"-fname"))
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

func compileTemplate(m map[string]Item) ([]byte, error) {
	items := make([]Item, 0, len(m))
	for _, v := range m {
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

			a:not(:last-child) > .file {
				border-bottom: 1px lightgray solid;
			}

			.file {
				margin: 10px;
				padding-bottom: 10px;
				display: flex;
			}

			.fname {
				width: 300px;
				margin-right: 10px;
				word-break: break-all;
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
		return []byte{}, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, items); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func updateTemplate() {
	if time.Now().Sub(lastUpdateTime).Seconds() >= 10 {
		var itemMap map[string]Item

		itemMap, err := readItems(dir, itemMap)
		if err != nil {
			panic(err)
		} else if l := len(itemMap); l != prevItemCount {
			log.Printf("read %d file(s)", l)

			tpl, err = compileTemplate(itemMap)
			if err != nil {
				panic(err)
			}
			prevItemCount = l
		}
	}
}

func authErr(w http.ResponseWriter, err string) {
	w.Header().Add("WWW-Authenticate", "Basic")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(err + "\n"))
}

func root(w http.ResponseWriter, r *http.Request) {
	if user, pass, ok := r.BasicAuth(); !ok {
		authErr(w, "auth required")
		return
	} else if correctUser != user || correctPass != pass {
		authErr(w, "incorrect user/pass")
		return
	}

	updateTemplate()

	w.Write(tpl)
}

func main() {
	// read auth information
	auth, err := ioutil.ReadFile(authfile)
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(auth), "\n")
	correctUser = strings.TrimSpace(lines[0])
	correctPass = strings.TrimSpace(lines[1])

	// set http handlers
	http.HandleFunc("/", root)
	log.Printf("listening on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
