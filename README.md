# ecs-restarter

## DESCRIPTION
CLI tool to discover ECS services by name substring across all clusters in an AWS region
You can select a matching service and choose to:
- Restart it (force new deployment with its current configuration)
- Scale desired task count (0–4)

Waits until the service reaches a stable state.

---

## USAGE

```bash
ecs-restarter -region <aws-region> -filter <substring> [-workers N]
```

## EXAMPLE

ecs-restarter -region eu-west-1 -filter myservice

## BUILD

```bash

for GOOS in linux darwin; do
    for GOARCH in amd64 arm64; do
        CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -o ecs-restarter.$GOOS.$GOARCH main.go
    done
done

# or

# ... assuming you've installed goreleaser
goreleaser release --snapshot --clean
```
