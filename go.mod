module github.com/openshift-kni/node-label-operator

go 1.16

require (
	github.com/go-logr/logr v0.3.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	sigs.k8s.io/controller-runtime v0.7.0
	sigs.k8s.io/controller-tools v0.0.0-00010101000000-000000000000
	sigs.k8s.io/kustomize/kustomize/v3 v3.0.0-00010101000000-000000000000
)

replace (
	k8s.io/client-go => k8s.io/client-go v0.19.2
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.4.1
	sigs.k8s.io/kustomize/kustomize/v3 => sigs.k8s.io/kustomize/kustomize/v3 v3.8.7
)
