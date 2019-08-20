module github.com/syseleven/designate-certmanager-webhook

go 1.12

require (
	github.com/gophercloud/gophercloud v0.0.0-20190126172459-c818fa66e4c8
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jetstack/cert-manager v0.8.0-alpha.0
	github.com/kubernetes-incubator/external-dns v0.5.15
	github.com/sirupsen/logrus v1.2.0
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20190413053546-d0acb7a76918
	k8s.io/client-go v11.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190413052642-108c485f896e

replace github.com/evanphx/json-patch => github.com/evanphx/json-patch v0.0.0-20190203023257-5858425f7550
