package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/Kunde21/athens-plugin/backend"
)

func main() {
	var plugin, dir string
	flag.StringVar(&plugin, "p", "testPlugin", "name of the plugin to run")
	flag.StringVar(&dir, "d", ".", "test athens directory")
	flag.Parse()

	ctx, canc := context.WithTimeout(context.Background(), 9*time.Second)
	defer canc()
	pl, err := backend.NewPlugin(ctx, plugin, "/tmp/storage.sock", dir)
	if err != nil {
		log.Fatal(err)
	}
	ctx, canc = context.WithTimeout(ctx, 2*time.Second)
	ex, err := pl.Exists(ctx, "cloud.google.com/go", "v0.26.0")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ex)
	inf, err := pl.Info(ctx, "cloud.google.com/go", "v0.26.0")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(inf))
	mod, err := pl.GoMod(ctx, "cloud.google.com/go", "v0.26.0")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%q\n", string(mod))
	n := time.Now()
	zip, err := pl.Zip(ctx, "cloud.google.com/go", "v0.26.0")
	if err != nil {
		log.Fatal(err)
	}
	defer zip.Close()
	f, err := os.Create("output.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fmt.Println(io.Copy(f, zip))
	fmt.Println(time.Since(n))
	n = time.Now()
	f.Seek(0, io.SeekStart)
	err = pl.Save(ctx, "testing.athens/plugin", "v0.0.2", mod, f, inf)
	if err != nil {
		log.Fatal("saving err", err)
	}
	fmt.Println(time.Since(n))
	ex, err = pl.Exists(ctx, "testing.athens/plugin", "v0.0.2")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ex)
}
