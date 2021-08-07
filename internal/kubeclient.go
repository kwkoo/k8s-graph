package internal

import (
	"context"
	"fmt"
	"log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	dynClient dynamic.Interface
}

func InitKubeClient(masterurl, kubeconfig string) (*KubeClient, error) {
	var (
		cfg *rest.Config
		err error
	)

	if masterurl != "" && kubeconfig != "" {
		if cfg, err = clientcmd.BuildConfigFromFlags(masterurl, kubeconfig); err != nil {
			return nil, err
		}
		log.Printf("initialized kube client with master URL %s and KUBECONFIG %s", masterurl, kubeconfig)
	}

	if cfg == nil {
		if cfg, err = rest.InClusterConfig(); err != nil {
			return nil, fmt.Errorf("could not initialize kube client using in-cluster config: %v", err)
		}
		log.Print("using in-cluster kube config")
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting new dynamic client: %v", err)
	}

	return &KubeClient{
		dynClient: dynClient,
	}, nil
}

func (kc *KubeClient) GetAll(ctx context.Context, namespace string) (Graph, error) {
	graph := InitGraph()
	if err := kc.GetPods(ctx, graph, namespace); err != nil {
		return *graph, err
	}

	return *graph, nil
}

func (kc *KubeClient) GetPods(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "", "v1", "pods", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "pod", item.GetName())
		for _, owner := range getOwners(item) {
			graph.addLink(owner, string(item.GetUID()))
		}
		// todo: process pod - container images, pvcs, configmaps, secrets
	}
	return nil
}

func (kc *KubeClient) get(ctx context.Context, g, v, r, namespace string) ([]unstructured.Unstructured, error) {
	resource := schema.GroupVersionResource{Group: g, Version: v, Resource: r}
	result, err := kc.dynClient.Resource(resource).Namespace(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func getOwners(u unstructured.Unstructured) []string {
	m, ok := u.Object["metadata"]
	if !ok {
		return nil
	}

	metadata, ok := m.(map[string]interface{})
	if !ok {
		return nil
	}
	or, ok := metadata["ownerReferences"]
	if !ok {
		return nil
	}
	ownerReferences, ok := or.([]interface{})
	if !ok {
		return nil
	}

	owners := []string{}
	for _, o := range ownerReferences {
		owner, ok := o.(map[string]interface{})
		if !ok {
			continue
		}
		ou, ok := owner["uid"]
		if !ok {
			continue
		}
		uid, ok := ou.(string)
		if !ok {
			continue
		}
		owners = append(owners, uid)
	}

	return owners
}
