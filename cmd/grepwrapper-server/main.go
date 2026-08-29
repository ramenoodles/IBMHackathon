package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"grepwrapper/internal/httpapi"
	"grepwrapper/internal/service"
)

func main() {
	root := flag.String(
		"root",
		".",
		"codebase directory to expose",
	)

	address := flag.String(
		"addr",
		":8080",
		"HTTP listen address",
	)

	rgBinary := flag.String(
		"rg",
		"rg",
		"path to ripgrep executable",
	)

	flag.Parse()

	lookupService, err := service.New(
		*root,
		*rgBinary,
	)
	if err != nil {
		log.Fatalf("create service: %v", err)
	}

	server := httpapi.New(lookupService)

	fmt.Printf(
		"grepwrapper API listening on %s\n",
		*address,
	)

	fmt.Printf(
		"search root: %s\n",
		*root,
	)

	if err := http.ListenAndServe(
		*address,
		server.Handler(),
	); err != nil {
		log.Fatal(err)
	}
}
