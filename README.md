# Graph

This is a web application that shows you the relationships between various resources deployed in an OpenShift project.

It uses the Kubernetes Go client in the backend, and [D3.js](https://d3js.org) and [Vue.js](https://vuejs.org) in the frontend.

![Screen Recording](images/graph.gif)


## Installation

To install this on OpenShift,

1. Edit `Makefile` and set the `PROJECT` variable to the namespace you want to this deploy to

1. Login as an administrator using the `oc` CLI

1. Run `make deploy`

To deploy this on non-OpenShift Kubernetes,

1. Edit `Makefile` and set the `PROJECT` variable to the namespace you want to this deploy to

1. Set the `KUBECONFIG` variable to point to your kubeconfig

1. Edit `yaml/k8s.yaml` and set the service `nodePort` to the relevant value

1. Run `make deploy-k8s` - a NodePort service is configured to listen on port 30080


## Todo

1. details screen

1. `configmaps` and `secrets` used in the `environment` in pod containers

1. `StatefulSets` and `DaemonSets`


## Resources

* [Unstructured docs](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured#Unstructured)