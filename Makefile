MASTERURL="https://192.168.254.10:6443"
KUBECONFIG="$(HOME)/.kube/config"
OPENSHIFT="false"
IMAGENAME="ghcr.io/kwkoo/k8s-graph"
PROJECT="graph"
VERSION="0.1"
BASE:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: run image deploy clean deploy-k8s clean-k8s

run:
	@MASTERURL="$(MASTERURL)" KUBECONFIG="$(KUBECONFIG)" OPENSHIFT="$(OPENSHIFT)" DOCROOT="$(BASE)/docroot" go run main.go

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

deploy-k8s:
	cat $(BASE)/yaml/k8s.yaml | sed 's|#PROJ#|$(PROJECT)|g' | KUBECONFIG=$(KUBECONFIG) kubectl apply -f -

clean-k8s:
	cat $(BASE)/yaml/k8s.yaml | sed 's|#PROJ#|$(PROJECT)|g' | KUBECONFIG=$(KUBECONFIG) kubectl delete -f -
