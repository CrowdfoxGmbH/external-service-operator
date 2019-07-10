package externalservice

import (
	"context"
	"github.com/CrowdfoxGmbH/external-service-operator/pkg/prober"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func runTestReconcile(client client.Client, name string, namespace string) (reconcile.Result, error) {
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	r := &ReconcileExternalService{client: client, scheme: scheme.Scheme, probeManager: prober.NewProber(client)}

	return r.Reconcile(req)
}

// Helper functions and short cuts

func getRuntimeEndpoint(client client.Client, name string, namespace string) (*corev1.Endpoints, error) {
	runtimeObject := &corev1.Endpoints{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, runtimeObject)

	return runtimeObject, err
}

func getRuntimeService(client client.Client, name string, namespace string) (*corev1.Service, error) {
	runtimeObject := &corev1.Service{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, runtimeObject)

	return runtimeObject, err
}

func createObject(client client.Client, obj runtime.Object) error {
	return client.Create(context.TODO(), obj)
}

func updateObject(client client.Client, obj runtime.Object) error {
	return client.Update(context.TODO(), obj)
}

func getRuntimeIngress(client client.Client, name string, namespace string) (*extv1.Ingress, error) {
	runtimeObject := &extv1.Ingress{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, runtimeObject)

	return runtimeObject, err
}
