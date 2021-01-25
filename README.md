Assignment
===

Create a Kuberentes Operator using the operator-sdk project that deploys wordpress using on sql via a custom resource.

## Expected results:
* You should be able to create a kubernetes resource of kind: Wordpress and apiVersion: example.com/v1 that will result in your operator deploying the Deployments, Secrets, PersistentVolumeClaims, Services, etc that consititute a simple instance of wordpress on sql.
* For a trivial example of exposing some configuration, allow the user to specify the plaintext password to that will end up in the Secret. (If you feel there is something better to expose as configuration, but do not go as far as replication)

## Example Usage:
* **User creates custom resource yaml:**

**wordpress.yaml**


```yaml
apiVersion: example.com/v1
kind: Wordpress
metadata:
  name: mysite
spec:
  sqlRootPassword: plaintextpassword
```

* **User deploys the yaml:**
```shell
$ kubectl create -f wordpress.yaml
wordpress.example.com/mysite created
```

* **Results:**
```shell
$ kubectl get deployment,service,pvc,secret
NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/wordpress         1/1     1            1           42s
deployment.apps/wordpress-mysql   1/1     1            1           42s

NAME                      TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)        AGE
service/kubernetes        ClusterIP      10.96.0.1       <none>          443/TCP        44m
service/wordpress         LoadBalancer   10.108.67.145   10.108.67.145   80:31735/TCP   42s
service/wordpress-mysql   ClusterIP      None            <none>          3306/TCP       42s

NAME                                   STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/mysql-pv-claim   Bound    pvc-323e96f8-e740-4b46-a22f-c28aa0aa3092   20Gi       RWO            standard       42s
persistentvolumeclaim/wp-pv-claim      Bound    pvc-ea2e3d46-65f3-456f-82f4-b7426786eb30   20Gi       RWO            standard       42s

NAME                           TYPE                                  DATA   AGE
secret/default-token-vmplq     kubernetes.io/service-account-token   3      2d9h
secret/mysql-pass-c57bb4t7mf   Opaque                                1      42s
```

## Useful imports:
```shell
import (
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

appsv1.Deployment{}
```
## Example Implementation:
Run this command to clone the example implementation and print the lines where modifications have been made to the boilerplate.

```shell
$ git clone git@github.com:operator-framework/operator-sdk-samples.git;cd operator-sdk-samples/go/memcached-operator;grep -Er 'TODO\(user\)|EDIT THIS' .
Cloning into 'operator-sdk-samples'...
remote: Enumerating objects: 102, done.
remote: Counting objects: 100% (102/102), done.
remote: Compressing objects: 100% (85/85), done.
remote: Total 16050 (delta 21), reused 83 (delta 12), pack-reused 15948
Receiving objects: 100% (16050/16050), 20.28 MiB | 1.35 MiB/s, done.
Resolving deltas: 100% (6443/6443), done.
./pkg/controller/memcached/memcached_controller.go:	// TODO(user): Modify this to be the types you create that are owned by the primary resource
./pkg/controller/memcached/memcached_controller.go:// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
./pkg/apis/cache/v1alpha1/memcached_types.go:// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
```

### Docs:

This is not comprehensive, you can find more examples, docs, etc by looking, but everything necessary to implement the operator (other than golang and kubernetes concepts) is here:

* The operator-sdk overview and quickstart provides step-by-step instructions.
* Project layout doc
* Example implementation referred to above.
* Optional: If you want to understand the Manager, Reconciler, Watch, etc. They use the controll-runtime library
### Env:
* Kubernetes on Minikube will be sufficient

### How to run:

```shell
make install
make run
```

### Deploy Sample Wordpress operator:
```shell
kubectl apply -f config/samples/wordpress_v1_wordpress.yaml
```

### Check if operator is deployed:
```shell
kubectl get deployment,svc,pv,service
```

### Run the deployment:
```shell
minikube service wordpress --url
```