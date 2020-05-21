module servicetrait

go 1.13

require (
	cloud.google.com/go v0.57.0 // indirect
	github.com/crossplane/crossplane-runtime v0.9.0
	github.com/crossplane/oam-controllers v0.0.0-00010101000000-000000000000
	github.com/crossplane/oam-kubernetes-runtime v0.0.1
	github.com/go-logr/logr v0.1.0
	github.com/google/go-cmp v0.4.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/onsi/ginkgo v1.12.2
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.6.0 // indirect
	github.com/prometheus/common v0.10.0 // indirect
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37 // indirect
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	k8s.io/api v0.18.3
	k8s.io/apiextensions-apiserver v0.18.3 // indirect
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200427153329-656914f816f9 // indirect
	k8s.io/utils v0.0.0-20200520001619-278ece378a50 // indirect
	sigs.k8s.io/controller-runtime v0.6.0
)

replace github.com/crossplane/oam-controllers => github.com/crossplane/addon-oam-kubernetes-local v0.0.0-20200519023759-42e82c49fb67
