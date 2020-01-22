package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"lsgooi/types"
	"lsgooi/webdav"
	"net/http"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"
)

const (
	port   = "8080"
	dir    = "/files"
	urlfmt = "https://f.lieuwe.xyz/vang/%s/%s"
)

var (
	state         State
	webdavHandler = webdav.MakeHandler(dir)
)

// readItems reads the given gooi files directory for items, if prev is non-nil
// it will be used as a cache for existing files. If a file is removed from the
// directory it isn't included in the result, even if prev does contain it.
func readItems(dir string, prev map[string]types.Item) (map[string]types.Item, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	m := make(map[string]types.Item)
	for _, f := range files {
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

		m[id] = types.Item{
			ID:   id,
			Name: strings.TrimSpace(string(fname)),
			Size: uint64(f.Size()),
			Date: f.ModTime(),
			URL:  fmt.Sprintf(urlfmt, id, fname),
		}
	}
	return m, nil
}

func compileTemplate(m map[string]types.Item) ([]byte, error) {
	items := make([]types.Item, 0, len(m))
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

			.file {
				margin: 10px;
				padding-bottom: 10px;
				display: flex;

				color: black;
				text-decoration: none;
			}

			.file:not(:last-child) {
				border-bottom: 1px lightgray solid;
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
			<a class="file" href="{{$v.URL}}">
				<div class="fname">{{$v.Name}}</div>
				<div class="size">{{$v.SizeString}}</div>
				<div class="date">{{$v.DateString}}</div>
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

// updateTemplate checks whether or not the template should be updated, and
// updates if necessary.
func updateTemplate() {
	if !state.MustCheck() {
		return
	}

	itemMap, err := readItems(dir, state.itemMap)
	if err != nil {
		panic(err)
	}

	state.lastCheckTime = time.Now()

	oldLen := len(state.itemMap)
	newLen := len(itemMap)
	if oldLen != newLen {
		log.Printf("read %d new file(s)", newLen-oldLen)

		state.itemMap = itemMap
		state.tpl, err = compileTemplate(itemMap)
		if err != nil {
			panic(err)
		}
		webdavHandler.Refresh(state.itemMap)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	updateTemplate()
	w.Write(state.tpl)
}

func main() {
	updateTemplate()

	// set http handlers
	http.HandleFunc("/", root)
	http.Handle("/webdav/", webdavHandler)
	log.Printf("listening on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
