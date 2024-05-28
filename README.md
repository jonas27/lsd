# List Secret Decode (LSD)

`lsd` is a tool to convert base64 decoded Kubernetes secrets to plain text secrets, either `yaml` or `json` format.

## Installation

### Go
```bash
```
go get github.com/jonas27/LSD
```

## Usage

LSD either accepts a single secret or a list of secrets in Kubernetes 
```
$ kubectl get secret <secret name> -o <yaml|json> | lsd
$ kubectl get secret -o <yaml|json> | lsd
$ lsd < kubectl get secret <secret name> <secret file>.<yaml|json>
```

## Example

```
$ ksd < secrets.json
```

> output
```json
{
    "apiVersion": "v1",
    "data": {
        "password": "secret",
        "app": "kubernetes secret decoder"
    },
    "kind": "Secret",
    "metadata": {
        "name": "kubernetes secret decoder",
        "namespace": "ksd"
    },
    "type": "Opaque"
}
```
