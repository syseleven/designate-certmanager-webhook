# ACME webhook Implementation for OpenStack Designate

This is an ACME webhook implementation for the [cert-manager](http://docs.cert-manager.io). It works with OpenStack Designate DNSaaS to generate certificates using DNS-01 challenges.

## Continuous integration

GitHub Actions runs two workflows:

- **Test Workflow** (`.github/workflows/test.yml`): On every pull request and push to `master`, the `check` job runs `make check`. On pushes to `master` only, the `test` job also runs `make test` with repository secrets and variables.
- **PR Integration Test** (`.github/workflows/pr-integration.yml`): After a successful **Test Workflow** run for a pull request, this workflow runs `make test` against the PR head commit. It uses the `openstack-integration` environment so a maintainer must approve before secrets are used (required for third-party fork PRs, where secrets are not available in the initial workflow).

### Repository setup (maintainers)

1. Create an environment named `openstack-integration` under **Settings → Environments**.
2. Enable **Required reviewers** and add the users or teams who may approve integration test runs.
3. Keep deployment branch rules permissive for this environment (or verify that `workflow_run` jobs are not blocked), since the job checks out the PR head SHA rather than only `master`.
4. Ensure repository **Actions** variables and secrets used by `make test` remain configured (same names as in the workflow files).
5. If `master` uses branch protection with required status checks, update them to the new job names: `check` (and optionally `test` from **Test Workflow** on push) and `test` from **PR Integration Test** for pull requests.

Fork contributors may also need **Approve workflow run** on the first workflow before **Test Workflow** runs; that is separate from approving the `openstack-integration` deployment for **PR Integration Test**.

## Prerequisites

To use this chart [Helm](https://helm.sh/) must be installed in your Kubernetes cluster. Setting up Kubernetes and Helm and is outside the scope of this README. Please refer to the Kubernetes and Helm documentation. You will also need the [cert-manager](https://github.com/cert-manager/cert-manager). Please refer to the cert-manager [documentation](https://docs.cert-manager.io) for full technical documentation for the project. This README assumes, the cert-manager is installed in the namespace `cert-manager`. Adapt examples accordingly, if you have installed it in a different namespace.

## Deployment

***Optional*** You can choose to pre-create your authentication secret or configure the values via helm. If you don't want to configure your credentials via helm, create a kubernetes secret in the cert-manager namespace (or where you have deployed cert-manager, with the SysEleven addon it is _syseleven-cert-manager_) containing your OpenStack credentials and the project ID with the DNS zone you would like to use:

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
  --from-literal=OS_APPLICATION_CREDENTIAL_NAME=<OpenStack Application Credential name> \
  --from-literal=OS_APPLICATION_CREDENTIAL_SECRET=<OpenStack Application Credential Secret value>
```

### Chart deployment

For now, we do not host a chart repository. To use this chart, you must clone this repository. Edit the values.yaml file and add your OpenStack settings if you did not create the secret before. Then you can install the helm chart with the command:

```
helm install --name designate-certmanager --namespace=cert-manager designate-certmanager-webhook
```

## Configuration

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

You are now ready to create you first certificate resource. The easiest way to accomplish this is to add an annotation to an Ingress rule. Please adapt this example for your own needs:

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
  - host: my.ingress.com
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
    - my.ingress.com
    secretName: myingress-cert
```
