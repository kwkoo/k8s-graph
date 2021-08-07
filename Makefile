MASTERURL="https://api.sandbox.x8i5.p1.openshiftapps.com:6443"
KUBECONFIG="/Users/kwkoo/.kube/config"
BASE:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: run

run:
	@MASTERURL="$(MASTERURL)" KUBECONFIG="$(KUBECONFIG)" DOCROOT="$(BASE)/docroot" go run main.go
