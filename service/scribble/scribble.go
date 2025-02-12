package scribble

import (
	"encoding/xml"
	"strconv"
)

const ScribbleMinThickness = 0.02
const ScribbleMaxThickness = 0.12

type GMA3 struct {
	XMLName     xml.Name `xml:"GMA3"`
	DataVersion string   `xml:"DataVersion,attr"`
	Scribble    Scribble `xml:"Scribble"`
}

type Scribble struct {
	XMLName xml.Name        `xml:"Scribble"`
	Name    string          `xml:"Name,attr"`
	Content ScribbleContent `xml:"Scribble"`
}

type ScribbleContent struct {
	XMLName xml.Name `xml:"Scribble"`
	Size    string   `xml:"Size,attr"`
	I       []string `xml:"I"`
}

func New(name string, paths []string) GMA3 {
	return GMA3{
		DataVersion: "2.2.1.1",
		Scribble: Scribble{
			Name: name,
			Content: ScribbleContent{
				Size: strconv.Itoa(len(paths)),
				I:    paths,
			},
		},
	}
}
