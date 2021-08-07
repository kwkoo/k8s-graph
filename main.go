package main

import (
	"context"
	"log"

	"github.com/kwkoo/k8s-graph/internal"
)

func main() {
	client, err := internal.InitKubeClient("https://api.sandbox.x8i5.p1.openshiftapps.com:6443", "/Users/kwkoo/.kube/config")
	if err != nil {
		log.Fatal(err)
	}
	graph, err := client.GetAll(context.Background(), "kwkoo-dev")
	if err != nil {
		log.Fatal(err)
	}
	log.Print(graph)
}
