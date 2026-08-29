package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"grepwrapper/internal/httpapi"
	"grepwrapper/internal/llm"
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

	model := flag.String(
		"model",
		"",
		"Watsonx model ID",
	)

	flag.Parse()

	var llmClient llm.Client

	apiKey := os.Getenv("WATSONX_API_KEY")
	projectID := os.Getenv("WATSONX_PROJECT_ID")

	if apiKey != "" && projectID != "" && *model != "" {
		client, err := llm.NewWatsonxClient(*model)
		if err != nil {
			log.Fatalf("create Watsonx client: %v", err)
		}

		llmClient = client

		log.Printf("Watsonx enabled with model %s", *model)
	} else {
		log.Println(
			"Watsonx disabled: WATSONX_API_KEY, WATSONX_PROJECT_ID, or -model not configured",
		)
	}

	lookupService, err := service.New(
		*root,
		*rgBinary,
		llmClient,
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
