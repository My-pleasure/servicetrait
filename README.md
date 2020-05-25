# How to use it locally
- Follow [this](https://github.com/crossplane/addon-oam-kubernetes-local) to install OAM Application Controller and OAM Core workload and trait controllers
- Install statefulsetworkload controller
```
cd $GOPATH/src/
git clone https://github.com/My-pleasure/statefulsetworkload.git
cd statefulsetworkload/

make docker-build IMG=<project-name>:tag  # eg:statefulsetworkload:v0.2
make deploy IMG=<project-name>:tag

# if you use kind to create kubernetes cluster, you should load IMG to kind cluster
kind load docker-image <project-name>:tag
```
- Run servicetrait controller
```
cd $GOPATH/src/
git clone -b v1alpha2 https://github.com/My-pleasure/servicetrait.git
cd servicetrait/

make install
make run
```
- Apply the examples
```
kubectl apply -f config/samples/
```
