module github.com/syseleven/designate-certmanager-webhook

go 1.12

require (
	github.com/gophercloud/gophercloud v0.12.0
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jetstack/cert-manager v0.9.1
	github.com/kubernetes-incubator/external-dns v0.5.15
	github.com/sirupsen/logrus v1.6.0
	k8s.io/client-go/v12 v12.0.0
)

replace k8s.io/client-go/v12 => /v12k8s.io/client-go v12.0.0

replace github.com/evanphx/json-patch => github.com/evanphx/json-patch v0.0.0-20190203023257-5858425f7550

replace git.apache.org/thrift.git => github.com/apache/thrift v0.13.0
