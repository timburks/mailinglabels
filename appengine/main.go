package main

import (
	"bufio"
	"github.com/jung-kurt/gofpdf"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
)

func init() {
	http.HandleFunc("/", RedirectHandler)
	http.HandleFunc("/labels", LabelsHandler)
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/index.html", 303)
}

type Address struct {
	Name         string
	Street       string
	CityStateZip string
}

func ReadAddresses(file io.Reader) (addresses []Address) {
	addresses = make([]Address, 0)
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		parts := strings.Split(line, "$")
		var address Address
		if len(parts) > 0 {
			address.Name = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 {
			address.Street = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			address.CityStateZip = strings.TrimSpace(parts[2])
		}
		addresses = append(addresses, address)
	}
	return
}

// Avery 5160 label formatter
func RenderLabels(pdffile io.Writer, addresses []Address) (err error) {
	count := 0

	pdf := gofpdf.New("P", "in", "Letter", "")

	pagew, pageh, _ := pdf.PageSize(0)

	labelh := 1.0                         // 1" high
	labelw := 2.5                         // 2 1/2" wide
	marginv := (pageh - 10*labelh) / 2.0  // 1/2" from top
	marginh := (pagew - 3.0*labelw) / 4.0 // label margin
	pdf.SetFont("Helvetica", "", 10)

	for _, address := range addresses {
		if count == 0 {
			pdf.AddPage()
		}
		row := count / 3
		col := count % 3
		x := marginh + float64(col)*(labelw+marginh)
		y := float64(row)*labelh + marginv
		pdf.SetXY(x, y)
		pdf.MultiCell(
			labelw,
			labelh/5,
			address.Name+"\n"+address.Street+"\n"+address.CityStateZip,
			"",    // no border
			"LM",  // left justify, middle
			false) // don't fill
		count += 1
		if count == 30 {
			count = 0
		}
	}
	pdf.Output(pdffile)
	return
}

func LabelsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	mediaType, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatal(err)
			}
			w.Header().Set("Content-Type", "application/pdf")
			addresses := ReadAddresses(p)
			RenderLabels(w, addresses)
		}
	} else {
		http.Redirect(w, r, "/index.html", 303)
	}
}
