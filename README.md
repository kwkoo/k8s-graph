# Graph

## Todo

1. `projects` (`project.openshift.io/v1`)

1. project selector

1. loading indicator

1. error screen

1. details screen

1. `configmaps` (`v1`)

1. `secrets` (`v1`)

1. `persistentvolumeclaims` (`v1`)


# Build

* `.status.output.to.imageDigest` set to `sha256:f6867f9cc7db9d27bad2b8af06f40f344a7014f4552152cc574670b76bc778d8`
* `.status.outputDockerImageReference` set to `image-registry.openshift-image-registry.svc:5000/kwkoo-dev/snake:latest`
* should create `Image` object with name `sha256:f6867f9cc7db9d27bad2b8af06f40f344a7014f4552152cc574670b76bc778d8` (equates to `ImageStreamTag`'s `.image.metadata.name`)
* in pod, `.spec.containers[0].image` set to `image-registry.openshift-image-registry.svc:5000/kwkoo-dev/snake@sha256:f6867f9cc7db9d27bad2b8af06f40f344a7014f4552152cc574670b76bc778d8`


## Resources

* [Unstructured docs](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured#Unstructured)