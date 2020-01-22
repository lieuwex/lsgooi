package main

import (
	"bytes"
	"log"
	"lsgooi/types"
	"lsgooi/webdav"
	"net/http"
	"sort"
	"text/template"
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

func root(w http.ResponseWriter, r *http.Request) {
	state.Update()
	w.Write(state.tpl)
}

func webdavfn(w http.ResponseWriter, r *http.Request) {
	state.Update()
	webdavHandler.ServeHTTP(w, r)
}

func main() {
	state.Update()

	// set http handlers
	http.HandleFunc("/", root)
	http.HandleFunc("/webdav/", webdavfn)
	log.Printf("listening on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
