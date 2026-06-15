package internal

import (
	"bytes"
	"io"
)

var ssePrefix = []byte("data: ")
var sseSuffix = []byte("\n")

func WriteSSE(w io.Writer, html []byte) error {
	for line := range bytes.SplitSeq(html, sseSuffix) {
		_, err := w.Write(ssePrefix)
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
	_, err := w.Write(sseSuffix)
	if err != nil {
		return err
	}
	return nil
}
