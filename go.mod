module github.com/syseleven/designate-certmanager-webhook

go 1.12

require (
	github.com/gophercloud/gophercloud v0.0.0-20190126172459-c818fa66e4c8
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jetstack/cert-manager v0.9.1
	github.com/kubernetes-incubator/external-dns v0.5.15
	github.com/sirupsen/logrus v1.2.0
	k8s.io/apiextensions-apiserver v0.0.0-20190718185103-d1ef975d28ce
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190413052642-108c485f896e

replace github.com/evanphx/json-patch => github.com/evanphx/json-patch v0.0.0-20190203023257-5858425f7550

replace github.com/jetstack/cert-manager => github.com/multi-io/cert-manager v0.9.1-test-propagationlimit-configurable
