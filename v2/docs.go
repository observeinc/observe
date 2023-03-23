package main

import "embed"

//go:embed *.md
//go:embed docs/*.md
var docFS embed.FS

func ReadDocFile(name string) ([]byte, error) {
	ret, err := docFS.ReadFile(name + ".md")
	if err != nil {
		ret, err = docFS.ReadFile("docs/" + name + ".md")
	}
	return ret, err
}
