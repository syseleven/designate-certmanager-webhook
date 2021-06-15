module github.com/syseleven/designate-certmanager-webhook

go 1.15

require (
	github.com/gophercloud/gophercloud v0.14.0
	github.com/jetstack/cert-manager v1.4.0
	github.com/kubernetes-incubator/external-dns v0.5.12
	github.com/sirupsen/logrus v1.7.0
	k8s.io/client-go v0.21.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.13.0
