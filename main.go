package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"transform/primitive"
)

var (
	IOCopyVar      = io.Copy
	tempFileVar    = tempfile
	listenAndServe = http.ListenAndServe
	tplExecuteBool = false
	TempName       = ""
)

type genOpts struct {
	N int
	M primitive.Mode
}

func main() {
	Controller()
}

func Controller() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", BasePath)
	mux.HandleFunc("/modify/", Modify)
	mux.HandleFunc("/upload", Upload)

	fs := http.FileServer(http.Dir("./img/"))
	mux.Handle("/img/", http.StripPrefix("/img/", fs))
	log.Fatal(listenAndServe(":3000", mux))
	return mux
}

// BasePath ...
func BasePath(w http.ResponseWriter, r *http.Request) {
	html := `<html>
		<form action="/upload" method="post" enctype="multipart/form-data">
		<body>	
		<input type="file" name="image"/>
			<button type="submit">Upload Image</button>
		<form/>
		</body>
		</html>`
	fmt.Fprint(w, html)
}

// Modify ...
func Modify(w http.ResponseWriter, r *http.Request) {
	//imgPath := r.URL.Path[len("/modify/"):]

	f, err := os.Open("./img/" + filepath.Base(r.URL.Path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer f.Close()
	ext := filepath.Ext(f.Name())
	modeStr := r.FormValue("mode")
	if modeStr == "" {
		renderModeChoices(w, r, f, ext)
		return
	}
	mode, err := strconv.Atoi(modeStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	nStr := r.FormValue("n")
	if nStr == "" {
		renderNumShapeChoices(w, r, f, ext, primitive.Mode(mode))
		return
	}
	numShapes, err := strconv.Atoi(nStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	_ = numShapes
	http.Redirect(w, r, "/img/"+filepath.Base(f.Name()), http.StatusFound)
	w.Header().Set("Content-Type", "image/png")
	IOCopyVar(w, f)
}

//Upload ...
func Upload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("No file found for the key specified : %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	ext := filepath.Ext(header.Filename)[1:]
	onDisk, err := tempFileVar("", ext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer onDisk.Close()
	_, err = IOCopyVar(onDisk, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/modify/"+filepath.Base(onDisk.Name()), http.StatusFound)
}

func renderNumShapeChoices(w http.ResponseWriter, r *http.Request, rs io.ReadSeeker, ext string, mode primitive.Mode) {
	opts := []genOpts{
		{
			N: 10,
			M: primitive.ModeCircle,
		},
		{
			N: 10,
			M: primitive.ModeEllipse,
		},
		{
			N: 10,
			M: primitive.ModeRect,
		},
		{
			N: 10,
			M: primitive.ModeBeziers,
		},
	}
	imgs, err := genImages(rs, ext, opts...)
	if err != nil {
		//panic(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := `<html><body>
		{{range .}}
		<a href="/modify/{{.Name}}?mode={{.Mode}}&n={{.NumShapes}}">
		<img style="width: 20%;" src="/img/{{.Name}}"/>
			</a>
		{{end}}
	</body></html>`

	tpl := template.Must(template.New("").Parse(html))
	type dataStruct struct {
		Name      string
		Mode      primitive.Mode
		NumShapes int
	}
	var data []dataStruct
	for i, img := range imgs {
		data := append(data, dataStruct{
			Name:      filepath.Base(img),
			Mode:      opts[i].M,
			NumShapes: opts[i].N,
		})
		err = tpl.Execute(w, data)

		if err != nil || tplExecuteBool {
			log.Println("tpl execution failed")
		}

	}
}
func renderModeChoices(w http.ResponseWriter, r *http.Request, rs io.ReadSeeker, ext string) {

	opts := []genOpts{
		{
			N: 10,
			M: primitive.ModeCircle,
		},
		{
			N: 10,
			M: primitive.ModeEllipse,
		},
		{
			N: 10,
			M: primitive.ModeRect,
		},
		{
			N: 10,
			M: primitive.ModeBeziers,
		},
	}

	imgs, err := genImages(rs, ext, opts...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := `<html><body>
		{{range .}}
		<a href="/modify/{{.Name}}?mode={{.Mode}}">
		<img style="width: 20%;" src="/img/{{.Name}}"/>
			</a>
		{{end}}
	</body></html>`

	tpl := template.Must(template.New("").Parse(html))
	type dataStruct struct {
		Name string
		Mode primitive.Mode
	}
	var data []dataStruct
	for i, img := range imgs {
		data := append(data, dataStruct{
			Name: filepath.Base(img),
			Mode: opts[i].M,
		})
		err = tpl.Execute(w, data)

		if err != nil || tplExecuteBool {
			log.Println("tpl execution failed")
		}

	}
}

func genImages(rs io.ReadSeeker, ext string, opts ...genOpts) ([]string, error) {

	var ret []string

	for _, opt := range opts {
		rs.Seek(0, 0)
		f, err := genImage(rs, ext, opt.N, opt.M)
		if err != nil {
			return nil, err
		}
		ret = append(ret, f)
	}
	return ret, nil

}

func genImage(r io.Reader, ext string, numShapes int, mode primitive.Mode) (string, error) {
	out, err := primitive.Transform(r, ext, numShapes, primitive.WithMode(primitive.ModeRotatedrect))
	if err != nil {
		return "", err
	}
	outFile, err := tempFileVar(TempName, ext)
	if err != nil {
		return "", err
	}
	defer outFile.Close()
	_, err = IOCopyVar(outFile, out)
	if err != nil {
		log.Println("IO Copy failed")
		return "", err
	}
	return outFile.Name(), nil
}
func tempfile(prefix, ext string) (*os.File, error) {
	//in, err := ioutil.Tempfile("","in_")
	in, err := ioutil.TempFile("./img/", prefix)
	if err != nil {
		return nil, errors.New("main: failed to create temp input file")
	}
	defer os.Remove(in.Name())
	return os.Create(fmt.Sprintf("%s.%s", in.Name(), ext))
}
