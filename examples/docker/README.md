# Docker Example

This example runs:

- `outwatch` as one container
- a placeholder Java runtime container as another container
- a bind-mounted `./generated` directory for dynamic Logback output

## Files

- `docker-compose.yml`: runtime wiring
- `../../Dockerfile`: project-level image build for outwatch
- `config/log-levels.yaml`: watched source config
- `app/logback.xml`: Logback config that includes generated logger overrides

## Setup

Create the generated output directory with appropriate permissions:

```bash
mkdir -p ./generated && chmod 777 ./generated
```

## Run

From this folder:

```bash
docker compose up --build
```

## Test live update

1. Edit `config/log-levels.yaml`.
2. Check generated file:

```bash
ls -la ./generated
cat ./generated/dynamic-loggers.xml
```

3. Confirm your Java app container is configured to use `/app-config/logback.xml`.

## Notes

- The `app` service is a placeholder to demonstrate mount and config wiring.
- Replace it with your actual Java image and startup command.
- Uses bind mount (`./generated`) instead of named volume for better control over permissions.
- Containers run as non-root for improved security; `./generated` is world-writable to allow writing from any UID.
