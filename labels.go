// Avery 5160 label formatter

package main

import (
	"bufio"
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"io"
	"os"
	"strings"
)

type Address struct {
	Name         string
	Street       string
	CityStateZip string
}

func ReadAddresses(filename string) (addresses []Address) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("unable to open address file\n")
		os.Exit(-1)
	}
	defer file.Close()

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
		address.Name = parts[0]
		address.Street = parts[1]
		address.CityStateZip = parts[2]
		addresses = append(addresses, address)
	}
	return
}

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

func main() {
	addresses := ReadAddresses("addresses.txt")
	pdffile, _ := os.Create("labels.pdf")
	RenderLabels(pdffile, addresses)
	pdffile.Close()
}
