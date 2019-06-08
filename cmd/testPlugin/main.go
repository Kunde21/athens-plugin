package main

import (
	"flag"
	"log"

	"github.com/Kunde21/athens-plugin/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/spf13/afero"
)

func main() {
	var config string
	flag.StringVar(&config, "config", "", "configuration plugin configuration")
	flag.Parse()
	st, err := fs.NewStorage(config, afero.NewOsFs())
	if err != nil {
		log.Fatal(err)
	}
	pl, err := storage.NewPlugin(st)
	if err != nil {
		log.Fatal(err)
	}
	defer pl.Close()
	if err := pl.Serve(); err != nil {
		log.Fatal(err)
	}
}
