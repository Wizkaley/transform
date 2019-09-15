package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"transform/primitive"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<html>
		<form action="/upload" method="post" enctype="multipart/form-data">
		<body>	
		<input type="file" name="image"/>
			<button type="submit">Upload Image</button>
		<form/>
		</body>
		</html>`
		fmt.Fprint(w, html)
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("image")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		ext := filepath.Ext(header.Filename)[1:]

		out, err := primitive.Transform(file, ext, 33, primitive.WithMode(primitive.ModeRotatedrect))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		outFile, err := tempfile("", ext)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer outFile.Close()
		io.Copy(outFile, out)
		redirUrl := fmt.Sprintf("/%s", outFile.Name())
		http.Redirect(w, r, redirUrl, http.StatusFound)

	})
	fs := http.FileServer(http.Dir("./img/"))
	mux.Handle("/img/", http.StripPrefix("/img/", fs))
	log.Fatal(http.ListenAndServe(":3000", mux))
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
