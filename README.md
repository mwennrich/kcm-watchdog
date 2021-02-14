# kcm-watchdog

***ALPHA version***

## Usage

```text
checking for broken kube-controller-manager and restart if needed

Usage:
  kcm-watchdog [command]

Available Commands:
  check       check for known problems
  help        Help about any command

Flags:
      --checkinterval duration   time between nodeReady checks (default 30s)
  -h, --help                     help for kcm-watchdog
      --kcm-max-checks int       number of checks till kcm is marked as failed (default 5)
```

## Example

```bash
kubectl apply -f deploy/kcm-watchdog.yaml
```
