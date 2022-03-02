# CITA-Cloud Operator

> **ATTENTIONS:** THE `MAIN` BRANCH MAY BE IN AN UNSTABLE OR EVEN BROKEN STATE DURING DEVELOPMENT.

## Overview

The CITA-Cloud Operator provides an easy and solid solution to deploy and manage a full CITA-Cloud blockchain service stack to the target [Kubernetes](https://kubernetes.io/) clusters in a scalable and high-available way. The CITA-Cloud Operator defines multiple custom resources on top of Kubernetes [Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/). The Kubernetes API can then be used in a declarative way to manage CITA-Cloud deployment stack and ensure its scalability and high-availability operation.

# Getting started
## Deploy CITA-Cloud operator

- Add CITA-Cloud operator repository
```shell
helm repo add cita-cloud-operator https://cita-cloud.github.io/cita-cloud-operator
```

- Create the namespace to install CITA-Cloud operator(if needed)
```shell
kubectl create ns cita
```

- Install CITA-Cloud operator
```shell
helm install cita-cloud-operator cita-cloud-operator/cita-cloud-operator -n=cita
```
If you want to enable the webhook function, you can use the following command, provided that the controller-manager has been installed in your k8s cluster
```shell
helm install cita-cloud-operator cita-cloud-operator/cita-cloud-operator --set enableWebhooks=true -n=cita 
```

- Verify the installation
```shell
kubectl get pod -ncita | grep cita-cloud-operator
```
The expected output is as follows:
```shell
NAME                                         READY   STATUS    RESTARTS   AGE
cita-cloud-operator-757fcf4466-gllsn         1/1     Running   0          45h
```

## Create a blockchain

- define a chain named `test-chain`
```shell
kubectl apply -f https://raw.githubusercontent.com/cita-cloud/cita-cloud-operator/master/config/samples/citacloud_v1_chainconfig.yaml
```

- create admin account
```shell
kubectl apply -f https://raw.githubusercontent.com/cita-cloud/cita-cloud-operator/master/config/samples/citacloud_v1_admin_account.yaml
```

- create three consensus accounts
```shell
kubectl apply -f https://github.com/cita-cloud/cita-cloud-operator/blob/master/config/samples/citacloud_v1_node1_account.yaml
kubectl apply -f https://github.com/cita-cloud/cita-cloud-operator/blob/master/config/samples/citacloud_v1_node2_account.yaml
kubectl apply -f https://github.com/cita-cloud/cita-cloud-operator/blob/master/config/samples/citacloud_v1_node3_account.yaml
```

- online the `test-chain`
```shell
kubectl patch chainconfig test-chain --patch '{"spec": {"action": "Online"}}' --type=merge -ncita
```

- initialize three consensus nodes
```shell
kubectl apply -f https://raw.githubusercontent.com/cita-cloud/cita-cloud-operator/master/config/samples/citacloud_v1_chainnode1.yaml
kubectl apply -f https://raw.githubusercontent.com/cita-cloud/cita-cloud-operator/master/config/samples/citacloud_v1_chainnode2.yaml
kubectl apply -f https://raw.githubusercontent.com/cita-cloud/cita-cloud-operator/master/config/samples/citacloud_v1_chainnode3.yaml
```

- start three consensus nodes
```shell
kubectl patch chainnode my-node-1 --patch '{"spec": {"action": "Start"}}' --type=merge -ncita
kubectl patch chainnode my-node-2 --patch '{"spec": {"action": "Start"}}' --type=merge -ncita
kubectl patch chainnode my-node-3 --patch '{"spec": {"action": "Start"}}' --type=merge -ncita
```

- you can see three pods running
```shell
kubectl get pod -ncita | grep my-node

NAME          READY   STATUS    RESTARTS   AGE
my-node-1-0   6/6     Running   0          4h
my-node-2-0   6/6     Running   0          4h
my-node-3-0   6/6     Running   0          4h
```



