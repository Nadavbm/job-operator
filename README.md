# job-operator

kubernetes operator built with [kubebuilder](https://book.kubebuilder.io/introduction.html)

the operator will create cronjob by using crd inside a namespace

### create operator with kubebuilder

init project:
```
kubebuilder init --domain example.com --repo example.com/job
```

create api:
```
kubebuilder create api --group cronjobs --version v1 --kind Job
```

edit and generate crd:
```
make generate
make manifests
```

push docker image:
```
make docker-build docker-push IMG="nadavbm/jobop:v0.0.1"
```

### testing operator with minikube

to test the operator, run minikube and the following commands:
```
sh test/cleanup.sh 
kubectl apply -f test/jobop.yaml 
kubectl apply -f test/testcrd.yaml
```

test crd changes with (connect to the relevant namespace with `kubens`):
```
k edit jobs.cronjobs.example.com jobop
```

