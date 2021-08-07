package internal

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	if err := kc.GetBuildConfigs(ctx, graph, namespace); err != nil {
		return *graph, err
	}

	if err := kc.GetBuilds(ctx, graph, namespace); err != nil {
		return *graph, err
	}

	if err := kc.GetDeployments(ctx, graph, namespace); err != nil {
		return *graph, err
	}

	if err := kc.GetReplicaSets(ctx, graph, namespace); err != nil {
		return *graph, err
	}

	if err := kc.GetPods(ctx, graph, namespace); err != nil {
		return *graph, err
	}

	// todo: services, routes

	// this is needed because d3.js doesn't like links pointing to nodes that
	// don't exist
	graph.cleanLinks()

	return *graph, nil
}

func (kc *KubeClient) GetBuildConfigs(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "build.openshift.io", "v1", "buildconfigs", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "buildconfig", item.GetName())
		addOwnerLinks(item, graph)
	}

	return nil
}

func (kc *KubeClient) GetBuilds(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "build.openshift.io", "v1", "builds", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "build", item.GetName())
		addOwnerLinks(item, graph)

		imageDigest := getString(item.Object, []string{"status", "output", "to", "imageDigest"})
		if imageDigest == "" {
			continue
		}
		colon := strings.LastIndex(imageDigest, ":")
		var uid string
		if colon == -1 {
			uid = imageDigest
		} else {
			uid = imageDigest[colon+1:]
		}
		graph.addNode(uid, "image", imageDigest)
		graph.addLink(string(item.GetUID()), uid)
	}

	return nil
}

func (kc *KubeClient) GetDeployments(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "apps", "v1", "deployments", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "deployment", item.GetName())
		addOwnerLinks(item, graph)
	}

	return nil
}

func (kc *KubeClient) GetReplicaSets(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "apps", "v1", "replicasets", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "replicaset", item.GetName())
		addOwnerLinks(item, graph)
	}

	return nil
}

func (kc *KubeClient) GetPods(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "", "v1", "pods", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "pod", item.GetName())
		addOwnerLinks(item, graph)

		// check if we need to link to container images
		containers := getList(item.Object, []string{"spec", "containers"})
		if len(containers) > 0 {
			for _, c := range containers {
				cm, ok := c.(map[string]interface{})
				if !ok {
					continue
				}
				image := getString(cm, []string{"image"})
				if image == "" {
					continue
				}
				sep := strings.LastIndex(image, "@sha256:")
				if sep == -1 {
					continue
				}
				graph.addLink(string(item.GetUID()), image[sep+len("@sha256:"):])
			}
		}

		// todo: process pod - pvcs, configmaps, secrets
	}

	return nil
}

func addOwnerLinks(u unstructured.Unstructured, graph *Graph) {
	for _, owner := range getOwners(u) {
		graph.addLink(owner, string(u.GetUID()))
	}
}

func (kc *KubeClient) GetProjects(ctx context.Context) ([]Project, error) {
	items, err := kc.get(ctx, "project.openshift.io", "v1", "projects", "")
	if err != nil {
		return nil, err
	}

	all := []Project{}

	for _, item := range items {
		metadata := getMap(item.Object, []string{"metadata"})
		if metadata == nil {
			continue
		}
		name := getString(metadata, []string{"name"})
		displayName := getString(metadata, []string{"annotations", "openshift.io/display-name"})
		if name == "" {
			continue
		}
		all = append(all, newProject(name, displayName))
	}

	return all, nil
}

func (kc *KubeClient) get(ctx context.Context, g, v, r, namespace string) ([]unstructured.Unstructured, error) {
	resource := schema.GroupVersionResource{Group: g, Version: v, Resource: r}

	var (
		result *unstructured.UnstructuredList
		err    error
	)
	if namespace == "" {
		result, err = kc.dynClient.Resource(resource).List(ctx, v1.ListOptions{})
	} else {
		result, err = kc.dynClient.Resource(resource).Namespace(namespace).List(ctx, v1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}
