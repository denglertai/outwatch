# Kubernetes Example

This example deploys a pod with two containers:

- `app`: placeholder Java runtime container using Logback
- `outwatch`: sidecar that watches config and writes generated Logback XML

Generated files are shared via an `emptyDir` volume mounted at `/generated`.

## Files

- `namespace.yaml`
- `configmap-outwatch-config.yaml`
- `configmap-app-logback.yaml`
- `deployment-with-sidecar.yaml`

## Apply

```bash
kubectl apply -f namespace.yaml
kubectl apply -f configmap-outwatch-config.yaml
kubectl apply -f configmap-app-logback.yaml
kubectl apply -f deployment-with-sidecar.yaml
```

## Verify

```bash
kubectl -n outwatch-example get pods
kubectl -n outwatch-example logs deploy/java-app-with-outwatch -c outwatch
kubectl -n outwatch-example exec deploy/java-app-with-outwatch -c outwatch -- ls -la /generated
kubectl -n outwatch-example exec deploy/java-app-with-outwatch -c outwatch -- cat /generated/dynamic-loggers.xml
```

## Notes

- Replace `ghcr.io/denglertai/outwatch:latest` with your published image.
- Replace the `app` container image/command with your real Java application.
- Update `configmap-outwatch-config.yaml` to change dynamic logger levels.

## Security: fsGroup 0

The deployment uses `securityContext.fsGroup: 0` to allow the non-root outwatch container (UID 65532) to write to the shared `emptyDir` volume:

- **What it does:** Kubernetes dynamically adds GID 0 (root group) to all process supplementary groups, making group-owned volumes writable.
- **Why:** The distroless image runs as UID 65532 for security; `fsGroup: 0` grants write access without running as root.
- **Requirement:** Volumes must have group-writable permissions (Kubernetes volume types like `emptyDir` have appropriate defaults).

This pattern is more secure than running containers as root and is the recommended approach for sidecar patterns in Kubernetes.
