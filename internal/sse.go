package internal

import (
	"bytes"
	"context"
	"io"

	"github.com/a-h/templ"
)

var ssePrefix = []byte("data: ")
var sseSuffix = []byte("\n")

func WriteSSE(w io.Writer, html []byte) error {
	var err error
	for line := range bytes.SplitSeq(html, sseSuffix) {
		_, err = w.Write(ssePrefix)
		if err != nil {
			return err
		}
		_, err = w.Write(line)
		if err != nil {
			return err
		}
		_, err = w.Write(sseSuffix)
		if err != nil {
			return err
		}
	}
	_, err = w.Write(sseSuffix)
	return err
}

func WriteTemplSSE(w io.Writer, component templ.Component) error {
	var err error
	_, err = w.Write(ssePrefix)
	if err != nil {
		return err
	}

	err = component.Render(context.TODO(), w)
	if err != nil {
		return err
	}

	_, err = w.Write(sseSuffix)
	return err
}
