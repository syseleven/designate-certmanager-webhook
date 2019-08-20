# ACME webhook Implementation for OpenStack Designate

This ACME webhook implementation for cert-manager works with OpenStack designate to generate certificates using DNS-01 challenge.

# Deployment
This webhook is deployed as a separate deployment and registered with kubernetes apiserver as and apiservice. For more details on how
to setup and extension api-server with kubernetes refer to the following [documentation](https://kubernetes.io/docs/tasks/access-kubernetes-api/setup-extension-api-server/#setup-an-extension-api-server-to-work-with-the-aggregation-layer)

All the deployment templates needed to setup this webhook can be generated using the following make command:
```
make rendered-manifest.yaml
```

And be applied to the cluster:
```
kubectl apply -f _out/rendered-manifest.yaml
```

# Configuration

To configure your issuer or cluster issuer to use this webhook as a dns01 solver use the following reference issuer template:

```
apiVersion: http://certmanager.k8s.io/v1alpha1
kind: Issuer
metadata:
  name: example-issuer
spec:
  acme:
    email: user@example.com
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: example-issuer-account-key
    solvers:
    - dns01:
        webhook:
          groupName: example.com
          solverName: name-of-solver
          config:
            Email: user@example.com
```
