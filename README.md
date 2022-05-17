# k8s simple logs utility

A simple log viewer for kubernetes hosted applications.

This tool exposes all the logs from every pod in the current namespace over http.

## Deploying
```
  curl -L https://raw.githubusercontent.com/fredsmith/k8s-simple-logs/main/k8s-deployment.yaml > k8s-simple-logs.yaml
  kubectl apply -f k8s-simple-logs.yaml
```

## Accessing

Add the service `logs` to your ingress like this:
```
          - path: /logs
            pathType: Prefix
            backend:
              service:
                name: logs
                port:
                  number: 8080
```
make sure it's before any less specific rules.
