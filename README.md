# myservice

## Description
The sample explains the steps to extend Kubernetes APIs by adding CRD and controller. 

## Prerequisites
- Install kubebuilder, please refer to [Quick Start](https://book.kubebuilder.io/quick-start.html).
- Install XCode on OSX with `xcode-select --install`.
- You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster. Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

## Create CRD and controller
1. Create a directory and initialize project 

```sh
mkdir ~/myservice && cd myservice
kubebuilder init --domain my.domain --repo my.domain/myservice
```

2. Create API with group and kind name

```sh
kubebuilder create api --group webapp --version v1 --kind MyService 
```

3. Create CRD and other relavent yaml

```sh
make manifests
```

4. Edit MyServiceSpec - api/v1/myservice_types.go, to add additional property in MyServiceSpec
```
type MyServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	DeploymentReplicas int32              `json:"deploymentReplicas"`
	DeploymentImage    string             `json:"deploymentImage"`
	ServiceType        corev1.ServiceType `json:"serviceType"`
	Command            []string           `json:"command,omitempty"`
	Args               []string           `json:"args,omitempty"`
}
```

5. Edit Reconcile function in controllers/myservice_controller.go, add details how it handles update of MyService objects

6. `make` to update CRD

7. Apply the CRD to Kubernetes cluster
```sh
make install 
```

8. Start the controller, it will connect to kube-apiserver. 
```sh
make run
```

## Test
1. Edit sample config/samples/webapp_v1_myservice.yaml, add mandatory properties - deploymentImage, serviceType
```
apiVersion: webapp.my.domain/v1
kind: MyService
metadata:
  labels:
    app.kubernetes.io/name: myservice
    app.kubernetes.io/instance: myservice-sample
    app.kubernetes.io/part-of: myservice
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: myservice
  name: myservice-sample
spec:
  deploymentReplicas: 3
  deploymentImage: nginx:latest
  serviceType: NodePort
```

2. Apply the MyService object - myservice-sample to cluster 
```sh
kubectl apply -f config/samples/webapp_v1_myservice.yaml
```

3. Confirm it is deployed and accessible. 
```sh
kubectl get node -owide
kubectl get service myservice-sample
curl http://<NodeIP>:<NodePort>
```

5. clean up MyService object and CRDs
```sh
kubectl delete -f config/samples/webapp_v1_myservice.yaml
make uninstall
```

## Deploy controller on K8s cluster
In above test section, the controller runs on local and connect to k8s cluster with kubeconfig credentials, in this section, the controller will be deployed on k8s cluster and connect access api-server with RBAC configuration. 

1. Prepare one image registry, for example Harbor. 

2. Configure image URL in Makefile
```
IMG ?= <IMAGE_REGISTRY>/><PATH>/controller:latest
```

2. Create and push docker image, please review details of Dockfile
```sh
make docker-build
make docker-push
```

3. Deploy CRDs, controller and depedencies to k8s cluster
```sh
make deploy
```

4. Test the controller with same step 1~3 in Test section. 

5. Udeploy controller, it will delete CRDs, controllers and other stuffs from the k8s cluster 

```sh
make undeploy
```
