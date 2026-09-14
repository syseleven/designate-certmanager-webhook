# ACME webhook Implementation for OpenStack Designate

This is designate-certmanager-webhook, an ACME webhook implementation for [cert-manager](http://docs.cert-manager.io).
It works with OpenStack Designate DNSaaS to generate certificates using DNS-01 challenges.

## Prerequisites

We recommend [Helm](https://helm.sh/) for installing designate-certmanager-webhook.
Setting up Kubernetes and Helm is outside the scope of this README.
You will also need [cert-manager](https://github.com/cert-manager/cert-manager).
Please refer to the cert-manager [documentation](https://docs.cert-manager.io) for full technical documentation on the project.

This README assumes that cert-manager is installed in the namespace `cert-manager`.
With the SysEleven addon it would be _syseleven-cert-manager_
Adapt examples accordingly if you have installed it in a different namespace.

The chart will be installed in the same namespace as cert-manager.

## Deployment

***Optional*** You can choose to pre-create your authentication secret or configure the values via helm.
If you don't want to configure your credentials via helm, create a kubernetes secret in the cert-manager namespace.

### Required permissions

The credential (user or application credential) must have the `member` role on the OpenStack project that owns the Designate zone. The webhook only performs these Designate operations, so `admin` is not required:

| HTTP call                                              | Used for                                   |
|--------------------------------------------------------|--------------------------------------------|
| `GET /v2/zones`                                        | Finding the zone for a challenge (by name) |
| `POST /v2/zones/{zone_id}/recordsets`                  | Publishing the DNS-01 TXT record           |
| `GET /v2/zones/{zone_id}/recordsets`                   | Locating the TXT record on cleanup         |
| `DELETE /v2/zones/{zone_id}/recordsets/{recordset_id}` | Removing the TXT record on cleanup         |

Zones are never created or deleted by the webhook. The project in the credential (`OS_PROJECT_ID`) must be the project that owns the zone.
Note that role names may differ between clouds; see [syseleven cloud](#permissions-in-syseleven-cloud) below.

### Permissions in SysEleven Cloud

The SysEleven cloud uses different access models depending on the region.
The Designate operations the webhook needs are listed in [Required permissions](#required-permissions) above.

**Regions `cbk` (Berlin), `dbl` (Berlin), `fes` (Frankfurt)**

These regions offer only two roles: `operator` (create, read, update, delete) and `viewer` (read-only).
The webhook needs the `operator` role on the project that owns the Designate zone; fine-grained or per-service roles are not available in these regions.

**Regions `ham1` (Hamburg) and `dus2` (Düsseldorf)**

The newer regions use the SysEleven cloud's [IAM](https://documentation.syseleven.de/en/products/iam/) with project-scoped, fine-grained permissions.
Create a [service account](https://documentation.syseleven.de/en/products/iam/usage/service-accounts/) and grant the following permissions:

- `can_become_dns_member_in_openstack`: write access isolated to the Designate DNS service, read access for listing zones and recordsets

All available permissions are documented in the [IAM authorization model](https://documentation.syseleven.de/en/products/iam/concepts/authorization-model/).

### Secret with OpenStack User Credentials

```
kubectl --namespace cert-manager create secret generic cloud-credentials \
  --from-literal=OS_AUTH_URL=<OpenStack Authentication URL> \
  --from-literal=OS_DOMAIN_NAME=<OpenStack Domain> \
  --from-literal=OS_REGION_NAME=<OpenStack Region> \
  --from-literal=OS_PROJECT_ID=<OpenStack Project ID> \
  --from-literal=OS_USERNAME=<OpenStack Username> \
  --from-literal=OS_PASSWORD=<OpenStack Password>
```

### Secret with OpenStack Application Credentials

```
kubectl --namespace cert-manager create secret generic cloud-credentials \
  --from-literal=OS_AUTH_URL=<OpenStack Authentication URL> \
  --from-literal=OS_DOMAIN_NAME=<OpenStack Domain> \
  --from-literal=OS_REGION_NAME=<OpenStack Region> \
  --from-literal=OS_APPLICATION_CREDENTIAL_ID=<OpenStack Application Credential ID> \
  --from-literal=OS_APPLICATION_CREDENTIAL_SECRET=<OpenStack Application Credential Secret value>
```

`OS_DOMAIN_NAME` will be "Default" in most installations.

### Chart deployment

The Helm chart is hosted inside in the Github repository, so there is no need to clone the repository if all you want to do it install the chart.

As mentioned above, the chart assumes that it is installed into the same namsepace as cert-manager.
This is because we need to set up RBAC to allow cert-manager to invoke the webhook, and it's easier to support that in the same namespace.
You could install the webhook in a different namespace, but then you'd have to create a matching clusterrolebinding separately.

To install the latest version of the chart:

```
helm repo add syseleven \
  https://syseleven.github.io/designate-certmanager-webhook

helm repo update

helm install designate-certmanager-webhook \
  syseleven/designate-certmanager-webhook \
  --namespace cert-manager
```

Pass `--version x.y.z` to install a specific chart version, and `--values myvalues.yaml` to override values.
Refer to ./helm/designate-certmanager-webhook/values.yaml to see which values can be set.

## Configuration

The chart registers the webhook as a Kubernetes [APIService](https://kubernetes.io/docs/reference/kubernetes-api/apiregistration/api-service-v1/) for consumption by cert-manager.

You need to create at least one DNS-01 Issuer or ClusterIssuer to make cert-manager aware of the webhook.

### Issuer

To configure your Issuer or ClusterIssuer to use this webhook as a DNS-01 solver use the following reference for a ClusterIssuer template. To use this in production please replace the reference to the Letsencrypt staging api accordingly:

```
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # You must replace this email address with your own.
    # Let's Encrypt will use this to contact you about expiring
    # certificates, and issues related to your account.
    email: user@example.com
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      # Secret resource used to store the account's private key.
      name: example-issuer-account-key
    # Add the designate dns webhook for dns challenges
    solvers:
    - dns01:
        webhook:
          groupName: acme.syseleven.de
          solverName: designatedns
```

## Usage

You are now ready to create your first certificate resource. The easiest way to accomplish this is to add an annotation to an Ingress rule. Please adapt this example for your own needs:

```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-staging
  name: myingress
spec:
  ingressClassName: nginx
  rules:
  - host: my.mydomain.com
    http:
      paths:
      - backend:
          service:
            name: myservice
            port:
              number: 1234
        path: /
        pathType: Prefix
  tls:
  - hosts:
    - my.mydomain.com
    secretName: mydomain-cert
```

Alternatively, you can also create the certificate resource directly.
That way you can also create wildcard certificates, which wouldn't be possible with HTTP-01 challenges:

```
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: mydomain-cert
spec:
  dnsNames:
  - '*.mydomain.com'
  issuerRef:
    group: cert-manager.io
    kind: ClusterIssuer
    name: letsencrypt-staging
  secretName: mydomain-cert
```
