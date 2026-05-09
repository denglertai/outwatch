# Docker Example

This example runs:

- `outwatch` as one container
- a placeholder Java runtime container as another container
- a shared `generated` volume for dynamic Logback output

## Files

- `docker-compose.yml`: runtime wiring
- `../../Dockerfile`: project-level image build for outwatch
- `config/log-levels.yaml`: watched source config
- `app/logback.xml`: Logback config that includes generated logger overrides

## Run

From this folder:

```bash
docker compose up --build
```

## Test live update

1. Edit `config/log-levels.yaml`.
2. Check generated file in volume:

```bash
docker compose exec outwatch ls -la /generated
docker compose exec outwatch cat /generated/dynamic-loggers.xml
```

3. Confirm your Java app container is configured to use `/app-config/logback.xml`.

## Notes

- The `app` service is a placeholder to demonstrate mount and config wiring.
- Replace it with your actual Java image and startup command.
