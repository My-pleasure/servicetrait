# ServiceTrait
You can use ServiceTrait to create a k8s service for workload on a Kubernetes cluster

Supported resources:
- ContainerizedWorkload
- StatefulSetWorkload
- StatefulSet
- Deployment
# Getting started
- At first, you should follow [addon-oam-kubernetes-local](https://github.com/crossplane/addon-oam-kubernetes-local). And install OAM Application Controller and OAM Core workload and trait controller.
- Then, you should deploy a StatefulSetWorkload controller by following [statefulsetworkload](https://github.com/My-pleasure/statefulsetworkload#getting-started).
- Get the Servicetrait project to your GOPATH
```
git clone https://github.com/My-pleasure/servicetrait.git
```
- Fetch the servicetrait image
```
docker pull chienwong/servicetrait:v1.1
```
- Deploy the servicetrait controller.
```
cd servicetrait/

make deploy IMG=chienwong/servicetrait:v1.1
```
- Apply the sample application config
```
kubectl apply -f config/samples/statefulsetworkload
```
- Verify ServiceTrait you should see a statefulset looking like below
```
kubectl get statefulset
NAME                         READY   AGE
example-appconfig-workload   1/1     19s
```
  And a service looking like below
```
kubectl get service
NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
example-appconfig-workload-svc   NodePort    10.103.191.25   <none>        80:32502/TCP   15s
```
# Support for K8S native resource
- Apply the sample application for K8S native statefulset
```
kubectl apply -f config/samples/statefulset/
```
- Verify ServiceTrait you should see a statefulset looking like below
```
kubectl get statefulset
NAME   READY   AGE
web    1/1     64s
```
  And a service looking like below
```
kubectl get service
NAME         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
web-svc      NodePort    10.101.142.126   <none>        80:30723/TCP   5s
```
