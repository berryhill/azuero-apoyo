# Azuero relief Helm chart

Adapted from the chart in `berryhill/elementa` at commit `e28747d` (the closest matching repository to “elemate”). This copies the Deployment, Service, ServiceAccount and optional Ingress patterns, not Elementa's Next.js application, MongoDB credentials or production release workflow.

The current directory contains a static HTML generator; it does not yet have a container image or publishing pipeline. Before deployment, build and publish an image that serves `index.html` on port 8080 as a non-root user, then set `image.repository` and an immutable `image.tag` or `image.digest`. The chart does not publish an image.

Default ingress is disabled because the Azuero relief domain/DNS/TLS secret are not established. To enable it, set `ingress.enabled=true`, `ingress.host`, and `ingress.tls.secretName` (or disable TLS explicitly). The cert-manager issuer and nginx ingress controller must already exist; adjust `ingress.annotations` to match the cluster. `existingEnvSecret` and `imagePullSecrets` reference pre-existing namespaced secrets only.

Persistence is enabled by default: the chart creates a 1Gi ReadWriteOnce PVC and mounts it read-write at `/data`. The single replica uses a Recreate rollout to avoid multi-attach conflicts (brief downtime during upgrades). The cluster must have a default StorageClass; otherwise set `persistence.storageClassName` or provide `persistence.existingClaim` for a claim you manage separately. For multiple replicas, provision ReadWriteMany-capable storage and set `persistence.accessModes={ReadWriteMany}`. Set `persistence.enabled=false` only if persistent data is not needed. The `/tmp` emptyDir remains ephemeral. A PVC does not populate data or make it web-accessible by itself; the container image must explicitly read or write `/data`. The chart-created PVC has Helm's `keep` annotation to retain it on uninstall; manage and back up that retained claim explicitly.

Offline validation:

    helm lint ./helm --set image.tag=test
    helm template azuero-relief ./helm --set image.tag=test

No cluster resources are applied by these checks. Do not run `helm upgrade --install` until the image, cluster target, and ingress settings are confirmed.
