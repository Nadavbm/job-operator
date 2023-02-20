#!/bin/sh
kubectl delete crd jobs.cronjobs.example.com
kubectl delete clusterrolebinding jobop
kubectl delete clusterrole jobop
kubectl delete ns jobop
kubectl delete ns job-test