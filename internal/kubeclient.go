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
	openShift bool
	dynClient dynamic.Interface
}

func InitKubeClient(masterurl, kubeconfig string, openShift bool) (*KubeClient, error) {
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
		openShift: openShift,
		dynClient: dynClient,
	}, nil
}

func (kc *KubeClient) GetAll(ctx context.Context, namespace string) (Graph, error) {
	graph := InitGraph()

	if err := kc.GetConfigMaps(ctx, graph, namespace); err != nil {
		return *graph, err
	}

	if err := kc.GetSecrets(ctx, graph, namespace); err != nil {
		return *graph, err
	}

	if err := kc.GetPersistentVolumeClaims(ctx, graph, namespace); err != nil {
		return *graph, err
	}

	if kc.openShift {
		if err := kc.GetBuildConfigs(ctx, graph, namespace); err != nil {
			return *graph, err
		}

		if err := kc.GetBuilds(ctx, graph, namespace); err != nil {
			return *graph, err
		}
	}

	if err := kc.GetDeployments(ctx, graph, namespace); err != nil {
		return *graph, err
	}

	if err := kc.GetStatefulSets(ctx, graph, namespace); err != nil {
		return *graph, err
	}

	if err := kc.GetDaemonSets(ctx, graph, namespace); err != nil {
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

	graph.cleanNodes()

	return *graph, nil
}

func (kc *KubeClient) GetBuildConfigs(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "build.openshift.io", "v1", "buildconfigs", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "buildconfig", item.GetName(), item.Object)
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
		graph.addNode(string(item.GetUID()), "build", item.GetName(), item.Object)
		addOwnerLinks(item, graph)

		imageDigest := getString(item.Object, "status", "output", "to", "imageDigest")
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
		graph.addNode(uid, "image", imageDigest, nil)
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
		graph.addNode(string(item.GetUID()), "deployment", item.GetName(), item.Object)
		addOwnerLinks(item, graph)
	}

	return nil
}

func (kc *KubeClient) GetStatefulSets(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "apps", "v1", "statefulsets", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "sts", item.GetName(), item.Object)
		addOwnerLinks(item, graph)
	}

	return nil
}

func (kc *KubeClient) GetDaemonSets(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "apps", "v1", "daemonsets", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "ds", item.GetName(), item.Object)
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
		graph.addNode(string(item.GetUID()), "replicaset", item.GetName(), item.Object)
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
		graph.addNode(string(item.GetUID()), "pod", item.GetName(), item.Object)
		addOwnerLinks(item, graph)

		// check if we need to link to container images
		containers := getList(item.Object, "spec", "containers")
		if len(containers) > 0 {
			for _, c := range containers {
				cm, ok := c.(map[string]interface{})
				if !ok {
					continue
				}
				image := getString(cm, "image")
				if image == "" {
					continue
				}
				sep := strings.LastIndex(image, "@sha256:")
				if sep == -1 {
					continue
				}
				graph.addLink(string(item.GetUID()), image[sep+len("@sha256:"):])

				// todo: check if we need to link configmaps or secrets used
				// in the environment
			}
		}

		volumes := getList(item.Object, "spec", "volumes")
		if len(volumes) > 0 {
			for _, v := range volumes {
				volume, ok := v.(map[string]interface{})
				if !ok {
					continue
				}
				claimName := getString(volume, "persistentVolumeClaim", "claimName")
				if claimName != "" {
					claimUid := graph.findResource("pvc", claimName)
					if claimUid == "" {
						continue
					}
					graph.addLink(string(item.GetUID()), claimUid)
					continue
				}
				cmName := getString(volume, "configMap", "name")
				if cmName != "" {
					cmUid := graph.findResource("cm", cmName)
					if cmUid == "" {
						continue
					}
					graph.addLink(string(item.GetUID()), cmUid)
					continue
				}
				secretName := getString(volume, "secret", "secretName")
				if secretName != "" {
					secretUid := graph.findResource("secret", secretName)
					if secretUid == "" {
						continue
					}
					graph.addLink(string(item.GetUID()), secretUid)
					continue
				}

			}
		}
	}

	return nil
}

func addOwnerLinks(u unstructured.Unstructured, graph *Graph) {
	for _, owner := range getOwners(u) {
		graph.addLink(owner, string(u.GetUID()))
	}
}

func (kc *KubeClient) GetPersistentVolumeClaims(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "", "v1", "persistentvolumeclaims", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "pvc", item.GetName(), item.Object)
	}

	return nil
}

func (kc *KubeClient) GetConfigMaps(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "", "v1", "configmaps", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "cm", item.GetName(), item.Object)
	}

	return nil
}

func (kc *KubeClient) GetSecrets(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "", "v1", "secrets", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "secret", item.GetName(), item.Object)
	}

	return nil
}

func (kc *KubeClient) GetProjects(ctx context.Context) ([]Project, error) {
	if !kc.openShift {
		// get namespaces
		return kc.getNamespaces(ctx)
	}

	// running against OpenShift
	items, err := kc.get(ctx, "project.openshift.io", "v1", "projects", "")
	if err != nil {
		return nil, err
	}

	all := []Project{}

	for _, item := range items {
		metadata := getMap(item.Object, "metadata")
		if metadata == nil {
			continue
		}
		name := getString(metadata, "name")
		displayName := getString(metadata, "annotations", "openshift.io/display-name")
		if name == "" {
			continue
		}
		all = append(all, newProject(name, displayName))
	}

	return all, nil
}

func (kc *KubeClient) getNamespaces(ctx context.Context) ([]Project, error) {
	items, err := kc.get(ctx, "", "v1", "namespaces", "")
	if err != nil {
		return nil, err
	}

	all := []Project{}

	for _, item := range items {
		metadata := getMap(item.Object, "metadata")
		if metadata == nil {
			continue
		}
		name := getString(metadata, "name")
		displayName := getString(metadata, "labels", "kubernetes.io/metadata.name")
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
