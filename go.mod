module github.com/syseleven/designate-certmanager-webhook

go 1.15

require (
	github.com/cert-manager/cert-manager v1.8.0
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/gophercloud/gophercloud v0.14.0
	github.com/kubernetes-incubator/external-dns v0.5.12
	github.com/sirupsen/logrus v1.8.1
	k8s.io/client-go v0.23.4
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.13.0
