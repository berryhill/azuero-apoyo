# Azuero relief Helm chart

Adapted from the chart in `berryhill/elementa` at commit `e28747d` (the closest matching repository to “elemate”). This copies the Deployment, Service, ServiceAccount and optional Ingress patterns, not Elementa's Next.js application, MongoDB credentials or production release workflow.

The Go API and Dockerfile serve the tracked `index.html` map and admin UI on port 8080. The database stays out of git. CI builds the committed page directly; deployment credentials are described in the root README.

The intended hostname is `azuero-apoyo.berryhill.dev` (configured in `ingress.host` with TLS secret `azuero-apoyo-tls`). Ingress remains disabled until DNS, the target cluster, and TLS issuer are confirmed; enable it with `ingress.enabled=true` when ready. The cert-manager issuer and nginx ingress controller must already exist; adjust `ingress.annotations` to match the cluster. `existingEnvSecret` and `imagePullSecrets` reference pre-existing namespaced secrets only.

The separate `secondaryIngress` routes `unidosporazuero.com` to the *same* Service and database as the existing domain, with its own cert-manager TLS secret `unidosporazuero-tls`. It is rendered when `ingress.enabled` and `secondaryIngress.enabled` are both true; it does not redirect or change the existing ingress. The apex currently serves a Netlify site. Before merging to the deploy branch, coordinate replacing its existing A records with the ingress controller address and confirm DNS resolution; this cutover will replace the Netlify site at the apex. Do not assume the ingress alone changes DNS. Once deployed, verify both HTTPS hostnames and the new certificate; this is a parallel domain cutover, not an isolated blue/green application deployment.

Persistence is enabled by default: the chart creates a 1Gi ReadWriteOnce PVC and mounts it read-write at `/data`. The single replica uses a Recreate rollout to avoid multi-attach conflicts (brief downtime during upgrades). The cluster must have a default StorageClass; otherwise set `persistence.storageClassName` or provide `persistence.existingClaim` for a claim you manage separately. For multiple replicas, provision ReadWriteMany-capable storage and set `persistence.accessModes={ReadWriteMany}`. Set `persistence.enabled=false` only if persistent data is not needed. The `/tmp` emptyDir remains ephemeral. A PVC does not populate data or make it web-accessible by itself; the container image must explicitly read or write `/data`. The chart-created PVC has Helm's `keep` annotation to retain it on uninstall; manage and back up that retained claim explicitly.

Offline validation:

    helm lint ./helm --set image.tag=test
    helm template azuero-relief ./helm --set image.tag=test

No cluster resources are applied by these checks. Do not run `helm upgrade --install` until the image, cluster target, and ingress settings are confirmed.
