# kcm-watchdog

***ALPHA version***

## Usage

```text
Usage:
  kcm-watchdog [command]

Available Commands:
  check       check for broken kube-controller-manager deployment and restart if needed
  help        Help about any command

Flags:
      --checkinterval duration   time between nodeReady checks (default 30s)
  -h, --help                     help for kcm-watchdog
      --kcm-max-fails int       number of checks till kcm is marked as failed (default 5)
```

## Example

```bash
kubectl apply -f deploy/kcm-watchdog.yaml
```
