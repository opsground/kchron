# kchron
// TODO(user): Add simple overview of use/purpose

## Description
// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started

### Prerequisites
- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/kchron:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/kchron:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/kchron:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/opsground/kchron/refs/heads/master/manifests/install.yaml
```

3. See the logs in `kchron-system` namespace

```sh
2024-10-05T00:41:11Z	INFO	Action	Setting up controller
2024-10-05T00:41:11Z	INFO	Action	Controller setup completed
2024-10-05T00:41:11Z	INFO	setup	starting manager
2024-10-05T00:41:11Z	INFO	controller-runtime.metrics	Starting metrics server
2024-10-05T00:41:11Z	INFO	setup	disabling http/2
2024-10-05T00:41:11Z	INFO	starting server	{"name": "health probe", "addr": "[::]:8081"}
I1005 00:41:11.038813       1 leaderelection.go:254] attempting to acquire leader lease kchron-system/3d0a3273.kchron.io...
I1005 00:41:11.111978       1 leaderelection.go:268] successfully acquired lease kchron-system/3d0a3273.kchron.io
2024-10-05T00:41:11Z	DEBUG	events	kchron-controller-manager-5cbdd79976-p29jc_9d8775c1-e05c-4933-8d5d-1d6a64f4bd82 became leader	{"type": "Normal", "object": {"kind":"Lease","namespace":"kchron-system","name":"3d0a3273.kchron.io","uid":"42e11a5b-d7be-46aa-98ce-c350b570346b","apiVersion":"coordination.k8s.io/v1","resourceVersion":"732032"}, "reason": "LeaderElection"}
2024-10-05T00:41:11Z	INFO	Starting EventSource	{"controller": "cronrestart", "controllerGroup": "resources.kchron.io", "controllerKind": "CronRestart", "source": "kind source: *v1alpha1.CronRestart"}
2024-10-05T00:41:11Z	INFO	Starting Controller	{"controller": "cronrestart", "controllerGroup": "resources.kchron.io", "controllerKind": "CronRestart"}
2024-10-05T00:41:11Z	INFO	Starting workers	{"controller": "cronrestart", "controllerGroup": "resources.kchron.io", "controllerKind": "CronRestart", "worker count": 1}
```

4. After creating CronRestart resource 

```sh
2024-10-05T00:41:11Z	INFO	Action	ObjectCreated	{"name": "guestbook-ui", "namespace": "default", "object": {"kind":"CronRestart","apiVersion":"resources.kchron.io/v1alpha1","metadata":{"name":"guestbook-ui","namespace":"default","uid":"d8d5eee9-4fa0-4ea6-b70c-431f8b1fe101","resourceVersion":"729767","generation":2,"creationTimestamp":"2024-10-05T00:05:38Z","labels":{"app.kubernetes.io/managed-by":"kustomize","app.kubernetes.io/name":"kchron"},"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"resources.kchron.io/v1alpha1\",\"kind\":\"CronRestart\",\"metadata\":{\"annotations\":{},\"labels\":{\"app.kubernetes.io/managed-by\":\"kustomize\",\"app.kubernetes.io/name\":\"kchron\"},\"name\":\"guestbook-ui\",\"namespace\":\"default\"},\"spec\":{\"cronSchedule\":\"*/3 * * * *\",\"namespace\":\"default\",\"resourceType\":\"Deployment\",\"resources\":[\"guestbook-ui\"]}}\n"},"managedFields":[{"manager":"kubectl-client-side-apply","operation":"Update","apiVersion":"resources.kchron.io/v1alpha1","time":"2024-10-05T00:12:28Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:kubectl.kubernetes.io/last-applied-configuration":{}},"f:labels":{".":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{}}},"f:spec":{".":{},"f:cronSchedule":{},"f:namespace":{},"f:resourceType":{},"f:resources":{}}}}]},"spec":{"namespace":"default","resourceType":"Deployment","resources":["guestbook-ui"],"cronSchedule":"*/3 * * * *"},"status":{}}}
2024-10-05T00:41:11Z	INFO	Action	Cron job scheduled	{"CronRestart": "guestbook-ui", "CronJobID": "cron-default-guestbook-ui"}
2024-10-05T00:41:11Z	INFO	controller-runtime.metrics	Serving metrics server	{"bindAddress": ":8443", "secure": true}
2024-10-05T00:42:00Z	INFO	Action	Successfully restarted Deployment	{"Namespace": "default", "Name": "guestbook-ui"}
```

5. After deleting the resource 

```sh
2024-10-05T00:47:20Z	INFO	Action	ObjectDeleted	{"name": "guestbook-ui", "namespace": "default"}
2024-10-05T00:47:20Z	INFO	Action	Cron job removed due to resource deletion	{"CronJobID": "cron-default-guestbook-ui"}
```


## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

