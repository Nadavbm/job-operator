### how to test

start minikube

cleanup if you already used minikube and deploy:
```
sh cleanup.sh
```

deploy gpDeployer to the cluster:
```
kubectl apply -f pgdeployer.yaml
```

test by applying new crd in another namespace:
```
kubectl apply -f testcrd.yaml
```

test by changing the crd (first connect to the relevant namespace with `kubens` or use `-n` for the relevant namespace):
```
kubectl edit pgdeployers.pgdeployer.example.com pgdeploy
```

