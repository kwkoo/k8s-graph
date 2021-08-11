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
		log.Printf("error getting ConfigMaps: %v", err)
	}

	if err := kc.GetSecrets(ctx, graph, namespace); err != nil {
		log.Printf("error getting Secrets: %v", err)
	}

	if err := kc.GetPersistentVolumeClaims(ctx, graph, namespace); err != nil {
		log.Printf("error getting PersistentVolumeClaims: %v", err)
	}

	if err := kc.GetCronJobs(ctx, graph, namespace); err != nil {
		log.Printf("error getting CronJobs: %v", err)
	}

	if err := kc.GetJobs(ctx, graph, namespace); err != nil {
		log.Printf("error getting Jobs: %v", err)
	}

	if err := kc.GetDeploymentConfigs(ctx, graph, namespace); err != nil {
		log.Printf("error getting DeploymentConfigs: %v", err)
	}

	if err := kc.GetBuildConfigs(ctx, graph, namespace); err != nil {
		log.Printf("error getting BuildConfigs: %v", err)
	}

	if err := kc.GetBuilds(ctx, graph, namespace); err != nil {
		log.Printf("error getting Builds: %v", err)
	}

	if err := kc.GetDeployments(ctx, graph, namespace); err != nil {
		log.Printf("error getting Deployments: %v", err)
	}

	if err := kc.GetStatefulSets(ctx, graph, namespace); err != nil {
		log.Printf("error getting StatefulSets: %v", err)
	}

	if err := kc.GetDaemonSets(ctx, graph, namespace); err != nil {
		log.Printf("error getting DaemonSets: %v", err)
	}

	if err := kc.GetReplicaSets(ctx, graph, namespace); err != nil {
		log.Printf("error getting ReplicaSets: %v", err)
	}

	if err := kc.GetReplicationControllers(ctx, graph, namespace); err != nil {
		log.Printf("error getting ReplicationControllers: %v", err)
	}

	if err := kc.GetPods(ctx, graph, namespace); err != nil {
		log.Printf("error getting Pods: %v", err)
	}

	if err := kc.GetServices(ctx, graph, namespace); err != nil {
		log.Printf("error getting Services: %v", err)
	}

	if err := kc.GetRoutes(ctx, graph, namespace); err != nil {
		log.Printf("error getting Routes: %v", err)
	}

	if err := kc.GetEndpointSlices(ctx, graph, namespace); err != nil {
		log.Printf("error getting EndpointSlices: %v", err)
	}

	// this is needed because d3.js doesn't like links pointing to nodes that
	// don't exist
	graph.cleanLinks()

	graph.cleanNodes()

	return *graph, nil
}

func (kc *KubeClient) GetCronJobs(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "batch", "v1beta1", "cronjobs", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "cj", item.GetName(), item.Object)
		addOwnerLinks(item, graph)
	}

	return nil
}

func (kc *KubeClient) GetJobs(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "batch", "v1", "jobs", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "job", item.GetName(), item.Object)
		addOwnerLinks(item, graph)
	}

	return nil
}

func (kc *KubeClient) GetBuildConfigs(ctx context.Context, graph *Graph, namespace string) error {
	if !kc.openShift {
		return nil
	}
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
	if !kc.openShift {
		return nil
	}
	items, err := kc.get(ctx, "build.openshift.io", "v1", "builds", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "build", item.GetName(), item.Object)
		addOwnerLinks(item, graph)

		imageDigest := unstructGetString(item.Object, "status", "output", "to", "imageDigest")
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

func (kc *KubeClient) GetDeploymentConfigs(ctx context.Context, graph *Graph, namespace string) error {
	if !kc.openShift {
		return nil
	}
	items, err := kc.get(ctx, "apps.openshift.io", "v1", "deploymentconfigs", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "dc", item.GetName(), item.Object)
		addOwnerLinks(item, graph)
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

func (kc *KubeClient) GetReplicationControllers(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "", "v1", "replicationcontrollers", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "rc", item.GetName(), item.Object)
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
		podid := string(item.GetUID())

		graph.addNode(podid, "pod", item.GetName(), item.Object)
		addOwnerLinks(item, graph)

		// check if we need to link to container images
		containers := unstructGetList(item.Object, "spec", "containers")
		if len(containers) > 0 {
			for _, c := range containers {
				cm, ok := c.(map[string]interface{})
				if !ok {
					continue
				}
				image := unstructGetString(cm, "image")
				if image == "" {
					continue
				}
				sep := strings.LastIndex(image, "@sha256:")
				if sep == -1 {
					continue
				}
				graph.addLink(podid, image[sep+len("@sha256:"):])

				// check for .spec.containers[*].envFrom
				ef := unstructGetList(cm, "envFrom")
				if len(ef) > 0 {
					for _, efitem := range ef {
						efitemmap, ok := efitem.(map[string]interface{})
						if !ok {
							continue
						}

						// check for .spec.containers[*].envFrom[*].configMapRef.name
						cmName := unstructGetString(efitemmap, "configMapRef", "name")
						if cmName != "" {
							uid := graph.findResource("cm", cmName)
							if uid == "" {
								continue
							}
							graph.addLink(podid, uid)
						} else {
							// check for .spec.containers[*].envFrom[*].secretRef.name
							secretName := unstructGetString(efitemmap, "secretRef", "name")
							if secretName != "" {
								uid := graph.findResource("secret", secretName)
								if uid == "" {
									continue
								}
								graph.addLink(podid, uid)
							}
						}
					}
				}

				// check for .spec.containers[*].env
				env := unstructGetList(cm, "env")
				if len(env) > 0 {
					for _, envItem := range env {
						envMap, ok := envItem.(map[string]interface{})
						if !ok {
							continue
						}

						// check for .spec.containers[*].env[*].valueFrom
						vf := unstructGetMap(envMap, "valueFrom")
						if vf == nil {
							continue
						}

						// check for .spec.containers[*].env[*].valueFrom.configMapKeyRef.name
						cmName := unstructGetString(vf, "configMapKeyRef", "name")
						if cmName != "" {
							uid := graph.findResource("cm", cmName)
							if uid == "" {
								continue
							}
							graph.addLink(podid, uid)
						} else {
							// check for .spec.containers[*].env[*].valueFrom.secretKeyRef.name
							secretName := unstructGetString(vf, "secretKeyRef", "name")
							if secretName != "" {
								uid := graph.findResource("secret", secretName)
								if uid == "" {
									continue
								}
								graph.addLink(podid, uid)
							}
						}
					}
				}
			}
		}

		volumes := unstructGetList(item.Object, "spec", "volumes")
		if len(volumes) > 0 {
			for _, v := range volumes {
				volume, ok := v.(map[string]interface{})
				if !ok {
					continue
				}
				claimName := unstructGetString(volume, "persistentVolumeClaim", "claimName")
				if claimName != "" {
					claimUid := graph.findResource("pvc", claimName)
					if claimUid == "" {
						continue
					}
					graph.addLink(podid, claimUid)
					continue
				}
				cmName := unstructGetString(volume, "configMap", "name")
				if cmName != "" {
					cmUid := graph.findResource("cm", cmName)
					if cmUid == "" {
						continue
					}
					graph.addLink(podid, cmUid)
					continue
				}
				secretName := unstructGetString(volume, "secret", "secretName")
				if secretName != "" {
					secretUid := graph.findResource("secret", secretName)
					if secretUid == "" {
						continue
					}
					graph.addLink(podid, secretUid)
					continue
				}

			}
		}
	}

	return nil
}

func addOwnerLinks(u unstructured.Unstructured, graph *Graph) {
	for _, owner := range unstructGetOwners(u) {
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

func (kc *KubeClient) GetServices(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "", "v1", "services", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		graph.addNode(string(item.GetUID()), "svc", item.GetName(), item.Object)
		addOwnerLinks(item, graph)
	}

	return nil
}

func (kc *KubeClient) GetRoutes(ctx context.Context, graph *Graph, namespace string) error {
	if !kc.openShift {
		return nil
	}
	items, err := kc.get(ctx, "route.openshift.io", "v1", "routes", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		uid := string(item.GetUID())
		graph.addNode(uid, "route", item.GetName(), item.Object)
		addOwnerLinks(item, graph)

		to := unstructGetMap(item.Object, "spec", "to")
		if to != nil {
			kind := unstructGetString(to, "kind")
			if kind == "Service" {
				name := unstructGetString(to, "name")
				if name != "" {
					svcuid := graph.findResource("svc", name)
					if svcuid != "" {
						graph.addLink(uid, svcuid)
					}
				}
			}
		}

		altBackends := unstructGetList(item.Object, "spec", "alternateBackends")
		if len(altBackends) > 0 {
			for _, b := range altBackends {
				backend, ok := b.(map[string]interface{})
				if !ok {
					continue
				}
				kind := unstructGetString(backend, "kind")
				if kind != "Service" {
					continue
				}
				name := unstructGetString(backend, "name")
				if name == "" {
					continue
				}
				svcuid := graph.findResource("svc", name)
				if svcuid == "" {
					continue
				}
				graph.addLink(uid, svcuid)
			}
		}
	}

	return nil
}

func (kc *KubeClient) GetEndpointSlices(ctx context.Context, graph *Graph, namespace string) error {
	items, err := kc.get(ctx, "discovery.k8s.io", "v1beta1", "endpointslices", namespace)
	if err != nil {
		return err
	}

	for _, item := range items {
		esuid := string(item.GetUID())
		graph.addNode(esuid, "endpointslice", item.GetName(), item.Object)
		addOwnerLinks(item, graph)

		endpoints := unstructGetList(item.Object, "endpoints")
		if len(endpoints) > 0 {
			for _, e := range endpoints {
				endpoint, ok := e.(map[string]interface{})
				if !ok {
					continue
				}
				targetRef := unstructGetMap(endpoint, "targetRef")
				if targetRef == nil {
					continue
				}
				kind := unstructGetString(targetRef, "kind")
				if kind != "Pod" {
					continue
				}
				podName := unstructGetString(targetRef, "name")
				if podName == "" {
					continue
				}
				poduid := graph.findResource("pod", podName)
				if poduid == "" {
					continue
				}
				graph.addLink(esuid, poduid)
			}
		}
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
		metadata := unstructGetMap(item.Object, "metadata")
		if metadata == nil {
			continue
		}
		name := unstructGetString(metadata, "name")
		displayName := unstructGetString(metadata, "annotations", "openshift.io/display-name")
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
		metadata := unstructGetMap(item.Object, "metadata")
		if metadata == nil {
			continue
		}
		name := unstructGetString(metadata, "name")
		displayName := unstructGetString(metadata, "labels", "kubernetes.io/metadata.name")
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
