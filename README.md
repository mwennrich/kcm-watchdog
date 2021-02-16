# shoot-watchdog

***ALPHA version***

## Usage

```text
Usage:
  shoot-watchdog [command]

Available Commands:
  check       check for broken shoot deployments and restart if needed
  help        Help about any command

Flags:
      --checkinterval duration    time between nodeReady checks (default 60s)
  -h, --help                      help for shoot-watchdog
      --shoot-max-fails int       number of checks till shoot is marked as failed (default 5)
```

## Example

```bash
kubectl apply -f deploy/shoot-watchdog.yaml
```
