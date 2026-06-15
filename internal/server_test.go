package internal

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/andybalholm/brotli"
	"testing"
)

var html = []byte(
	`<div id="fighter-sides"><div id="left-fighter-side"><div id="FighterInnerJQuery"><div>JQuery</div><div>Health: <ul id=""><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li></ul></div><div>Damage: 4</div><div>Speed 8</div><div>Timer: <ul id=""><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li></ul></div><div>Accuracy: 50% </div><div>Crit: 0% </div></div><div id="left-fighter-icon" class="animate-attack-left"><?xml version="1.0" encoding="utf-8"?><!-- Uploaded to: SVG Repo, www.svgrepo.com, Generator: SVG Repo Mixer Tools -->
<svg width="800px" height="800px" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
<circle cx="16" cy="16" r="14" fill="#0769AD"/>
<path d="M22.6573 13.4211C23.9143 13.4211 25.0652 13.0019 25.955 12.3066C25.0312 13.5238 23.5007 14.3196 21.7689 14.3196C18.9477 14.3196 16.6607 12.2077 16.6607 9.60256C16.6607 8.1581 17.3638 6.86527 18.4712 6C17.8901 6.76568 17.5491 7.6981 17.5491 8.70407C17.5491 11.3092 19.8361 13.4211 22.6573 13.4211Z" fill="#78CFF5"/>
<path d="M25.9064 16.6586C24.5512 17.7216 22.7968 18.3628 20.8805 18.3628C16.5874 18.3628 13.1071 15.1447 13.1071 11.1749C13.1071 9.63522 13.6307 8.20859 14.5221 7.03894C12.8413 8.35742 11.7745 10.3248 11.7745 12.5226C11.7745 16.4924 15.2548 19.7106 19.5479 19.7106C22.176 19.7106 24.4994 18.5047 25.9064 16.6586Z" fill="#78CFF5"/>
<path d="M26 20.7701C24.0689 22.6129 21.3937 23.7538 18.4375 23.7538C12.5497 23.7538 7.77678 19.2283 7.77678 13.6458C7.77678 11.8768 8.25603 10.214 9.09813 8.76767C7.18322 10.595 6 13.1125 6 15.892C6 21.4745 10.7729 26 16.6607 26C20.6827 26 24.1846 23.8881 26 20.7701Z" fill="#78CFF5"/>
</svg></div></div><div id="right-fighter-side"><div id="right-fighter-icon" class="animate-defend-right"><svg xmlns="http://www.w3.org/2000/svg" width="200" height="200" viewBox="0 0 256 168">
  <path fill="#111" d="M181.395 42.749L256 74.204v21.858l-74.605 31.017l-5.917-21.497l55.169-20.705l-55.169-20.784zm-106.79-.001L0 74.204v21.858l74.605 31.017l5.917-21.497l-55.169-20.705l55.169-20.784z"/>
  <path fill="#4065C5" d="M144.34 0h25.664L112.99 167.111H85.996z"/>
</svg>
</div><div id="FighterInnerHTMX"><div>HTMX</div><div>Health: <ul id=""><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li></ul></div><div>Damage: 10</div><div>Speed 8</div><div>Timer: <ul id=""><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li><li></li></ul></div><div>Accuracy: 99% </div><div>Crit: 40% </div></div></div></div><ul id="eventlog"><li>HTMX just hit JQuery for 4</li><li>JQuery just missed...</li><li>HTMX just hit JQuery for 4</li><li>JQuery just hit HTMX for 4</li></ul>`,
)

func BenchmarkGzip(b *testing.B) {
	var buf bytes.Buffer
	buf.Grow(300)
	w, err := gzip.NewWriterLevel(&buf, 5)
	if err != nil {
		fmt.Printf("err creating gzip writer: %v", err)
	}
	defer func() {
		if err = w.Close(); err != nil {
			fmt.Printf("err closing: %v\n", err)
		}
	}()
	for b.Loop() {
		err = WriteSSE(w, html)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		err = w.Flush()
		if err != nil {
			fmt.Printf("error flushing writer %v", err)
		}
		sampleLen := float64(len(html))
		compressedLen := float64(buf.Len())
		ratio := sampleLen / compressedLen
		// diff := sampleLen - compressedLen
		b.ReportMetric(sampleLen, "originalbytes")
		b.ReportMetric(compressedLen, "compressedbytes")
		b.ReportMetric(ratio, "ratio")

		buf.Reset()
	}
}
func BenchmarkBrotli(b *testing.B) {
	var buf bytes.Buffer
	buf.Grow(300)
	w := brotli.NewWriterOptions(&buf, brotli.WriterOptions{
		LGWin:   24,
		Quality: 3,
	})

	defer func() {
		if err := w.Close(); err != nil {
			fmt.Printf("err closing: %v\n", err)
		}
	}()

	for b.Loop() {
		err := WriteSSE(w, html)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		err = w.Flush()
		if err != nil {
			fmt.Printf("error flushing writer %v", err)
		}

		sampleLen := float64(len(html))
		compressedLen := float64(buf.Len())
		ratio := sampleLen / compressedLen
		// diff := sampleLen - compressedLen
		b.ReportMetric(sampleLen, "originalbytes")
		b.ReportMetric(compressedLen, "compressedbytes")
		b.ReportMetric(ratio, "ratio")

		buf.Reset()
	}

}
