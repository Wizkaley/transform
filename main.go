package main

import (
	"fmt"
	"io"
	"os"
	"transform/primitive"
)

func main() {

	file, err := os.Open("rnm.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	out, err := primitive.Transform(file, 50)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(string(out))

	os.Remove("out.png")
	f, err := os.Create("out.png")
	if err != nil {
		panic(err)
	}
	io.Copy(f, out)
}
