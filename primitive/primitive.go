package primitive

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Mode int

//Modes supported by the primitive package
const (
	ModeCombo Mode = iota
	ModeTriangle
	ModeRect
	ModeEllipse
	ModeCircle
	ModeRotatedrect
	ModeBeziers
	ModeRotatedellipse
	ModePolygon
)

// WithMode is an option for the transform function that
// will define the mode you want to use. By Default, ModeTriangle
// will be used.
func WithMode(mode Mode) func() []string {
	return func() []string {
		return []string{"-n", fmt.Sprintf("%d", mode)}
	}
}

// Transform will take the provided image and apply a primitivw
// transformation to it , then return a reader to the resultingn image
func Transform(image io.Reader, numShapes int, options ...func() []string) (io.Reader, error) {
	in, err := tempfile("_in", "png")
	if err != nil {
		return nil, errors.New("primitive: failed to create temp input file")
	}
	defer os.Remove(in.Name())
	out, err := tempfile("_in", "png")
	if err != nil {
		return nil, errors.New("primitive: failed to create temp output file")
	}
	defer os.Remove(out.Name())

	_, err = io.Copy(in, image)

	if err != nil {
		return nil, errors.New("primitive: failed to copy image into temp input file")
	}

	stdCombo, err := primitive(in.Name(), out.Name(), numShapes, ModeCombo)
	if err != nil {
		return nil, fmt.Errorf("primitive: failed to run the primitive command")
	}
	fmt.Println(stdCombo)
	b := bytes.NewBuffer(nil)
	_, err = io.Copy(b, out)
	if err != nil {
		return nil, errors.New("primitive: failed to copy output file into buffer")
	}
	return b, nil
}

func primitive(inputFile, outputFile string, numShape int, mode Mode) (string, error) {
	args := fmt.Sprintf("-i %s -o %s -n %d -m %b", inputFile, outputFile, numShape, mode)
	c := exec.Command("primitive", strings.Fields(args)...)
	b, err := c.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	return string(b), err
}

func tempfile(prefix, ext string) (*os.File, error) {
	//in, err := ioutil.Tempfile("","in_")
	in, err := ioutil.TempFile("", "in_")
	if err != nil {
		return nil, errors.New("primitive: failed to create temp input file")
	}
	defer os.Remove(in.Name())
	return os.Create(fmt.Sprintf("%s.%s", in.Name(), ext))
}
