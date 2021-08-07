package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kwkoo/configparser"
	"github.com/kwkoo/k8s-graph/internal"
)

//go:embed docroot/*
var content embed.FS

var client *internal.KubeClient

func graphHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	graph, err := client.GetAll(context.Background(), "kwkoo-dev")
	if err != nil {
		writeJSON(w, struct {
			Err string `json:"error"`
		}{
			Err: err.Error(),
		})
		return
	}
	writeJSON(w, graph)
}

func main() {
	config := struct {
		Port       int    `default:"8080" usage:"HTTP listener port"`
		Docroot    string `usage:"HTML document root - will use the embedded docroot if not specified"`
		MasterURL  string `usage:"Kubernetes master URL - will use the in-cluster config if not specified"`
		Kubeconfig string `usage:"Path to the kubeconfig file - will use the in-cluster config if not specified"`
	}{}
	if err := configparser.Parse(&config); err != nil {
		log.Fatal(err)
	}

	var filesystem http.FileSystem
	if len(config.Docroot) > 0 {
		log.Printf("using %s in the file system as the document root", config.Docroot)
		filesystem = http.Dir(config.Docroot)
	} else {
		log.Print("using the embedded filesystem as the docroot")

		subdir, err := fs.Sub(content, "docroot")
		if err != nil {
			log.Fatalf("could not get subdirectory: %v", err)
		}
		filesystem = http.FS(subdir)
	}

	fileServer := http.FileServer(filesystem).ServeHTTP

	var err error
	client, err = internal.InitKubeClient(config.MasterURL, config.Kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// Setup signal handling.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", config.Port),
	}
	go func() {
		log.Printf("listening on port %v", config.Port)
		http.HandleFunc("/api/graph", graphHandler)
		http.HandleFunc("/", fileServer)
		wg.Add(1)
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Print("web server graceful shutdown")
				return
			}
			log.Fatal(err)
		}
	}()

	// Wait for SIGINT
	<-shutdown
	log.Print("initiating web server shutdown...")
	signal.Reset(os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	wg.Wait()
	log.Print("shutdown successful")
	/*
		graph, err := client.GetAll(context.Background(), "kwkoo-dev")
		if err != nil {
			log.Fatal(err)
		}
		log.Print(graph)
	*/
}

func writeJSON(w io.Writer, data interface{}) {
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		log.Printf("error converting data to JSON: %v", err)
	}
}
