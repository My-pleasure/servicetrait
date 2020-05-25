# ServiceTrait
Run OAM ServiceTrait on a Kubernetes cluster, like ManualScalerTrait.
# How to use it
- Install OAM Application Controller and OAM Core workload and trait controller. (You can also follow [addon-oam-kubernetes-local](https://github.com/crossplane/addon-oam-kubernetes-local))
```
kubectl create namespace crossplane-system

helm repo add crossplane-alpha https://charts.crossplane.io/alpha

helm install crossplane --namespace crossplane-system crossplane-alpha/crossplane

git clone git@github.com:crossplane/addon-oam-kubernetes-local.git

kubectl create namespace oam-system

helm install controller -n oam-system ./charts/oam-core-resources/ 
```
- Install statefulsetworkload controller.
```
git clone https://github.com/My-pleasure/statefulsetworkload.git

cd $GOPATH/src/statefulsetworkload

make docker-build IMG=<project-name>:tag  # eg:statefulsetworkload:v0.2

make deploy IMG=<project-name>:tag

# if you use kind to create kubernetes cluster, you should load IMG to kind cluster
kind load docker-image <project-name>:tag
```
- Run servicetrait controller.
```
git clone -b v1alpha2 https://github.com/My-pleasure/servicetrait.git

cd $GOPATH/src/servicetrait

make install

make run
```
- Apply the samples
```
kubectl apply -f config/samples/
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
NAME                         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
example-appconfig-workload   ClusterIP   10.107.41.172   <none>        80/TCP    23s
```
