# Security Policy

## Reporting Security Issues

If you discover a security vulnerability in outwatch, please create an issue:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will:
1. Acknowledge receipt within 48 hours
2. Confirm the issue and assess severity
3. Develop and test a fix
4. Release a patched version
5. Credit you in release notes (if desired)

## Supported Versions

| Version | Status | Support End |
|---------|--------|-------------|
| 0.1.x   | Current | TBD |

## Security Best Practices

When deploying outwatch:
- ✅ Run as non-root user (container uses `nonroot`)
- ✅ Use read-only filesystems where possible
- ✅ Validate input YAML before mounting to pod
- ✅ Use network policies to restrict pod communication
- ✅ Keep dependencies updated via Renovate
- ✅ Review release notes for security patches
- ✅ Enable Pod Security Standards in Kubernetes
