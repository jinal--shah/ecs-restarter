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
# Assumes you have relevant AWS creds in env or .aws or usual places
ecs-restarter -region <aws-region> -filter <substring> [-workers N]
```

## EXAMPLE

ecs-restarter -region eu-west-1 -filter myservice

## BUILD LOCALLY

```bash
# ... assuming you've installed goreleaser
goreleaser release --snapshot --clean
```
