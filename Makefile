#MASTERURL="https://api.sandbox.x8i5.p1.openshiftapps.com:6443"
MASTERURL="https://api.cluster-ee7c.ee7c.sandbox1471.opentlc.com:6443/"
KUBECONFIG="/Users/kwkoo/.kube/config"
IMAGENAME="ghcr.io/kwkoo/k8s-graph"
PROJECT="graph"
VERSION="0.1"
BASE:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: run image deploy clean

run:
	@MASTERURL="$(MASTERURL)" KUBECONFIG="$(KUBECONFIG)" DOCROOT="$(BASE)/docroot" go run main.go

image:
	docker build --rm -t $(IMAGENAME):$(VERSION) $(BASE)
	docker tag $(IMAGENAME):$(VERSION) $(IMAGENAME):latest
	docker push $(IMAGENAME):$(VERSION)
	docker push $(IMAGENAME):latest

deploy:
	cat $(BASE)/yaml/openshift.yaml | sed 's|#PROJ#|$(PROJECT)|g' | oc apply -f -
	@echo "http://`oc get route/k8s-graph -n $(PROJECT) -o jsonpath='{.spec.host}'`"

clean:
	cat $(BASE)/yaml/openshift.yaml | sed 's|#PROJ#|$(PROJECT)|g' | oc delete -f -
