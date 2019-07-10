package testutils

import (
	esov1alpha1 "github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/api/extensions/v1beta1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func ExpectEqStr(actual string, expected string, t *testing.T) {
	if actual != expected {
		t.Errorf("Expected String '%v' but got '%v'", expected, actual)
	}
}

func ExpectTrue(result bool, t *testing.T) {
	if !result {
		t.Errorf("Expected bool to be true but got false")
	}
}

func ExpectFalse(result bool, t *testing.T) {
	if result {
		t.Errorf("Expected bool to be false but got true")
	}
}

func ExpectServiceType(actual corev1.ServiceType, expected corev1.ServiceType, t *testing.T) {
	if actual != expected {
		t.Errorf("Expected ServiceType '%v' but got '%v'", expected, actual)
	}
}

func ExpectEqInt(actual int32, expected int32, t *testing.T) {
	if actual != expected {
		t.Errorf("Expected int32 '%v' but got '%v'", expected, actual)
	}
}

func ExpectEndpointOwnerReference(actual *corev1.Endpoints, instance *esov1alpha1.ExternalService, t *testing.T) {
	if len(actual.ObjectMeta.OwnerReferences) != 1 {
		t.Fatalf("Ownerreference missing: %v", actual)
	}

	if actual.ObjectMeta.OwnerReferences[0].Kind != "ExternalService" {
		t.Fatalf("Ownerreference Kind does not match: %v", actual.ObjectMeta.OwnerReferences[0].Kind)
	}

	if actual.ObjectMeta.OwnerReferences[0].APIVersion != "eso.crowdfox.com/v1alpha1" {
		t.Fatalf("Ownerreference API does not match: %v", actual.ObjectMeta.OwnerReferences[0].APIVersion)
	}

	if actual.ObjectMeta.OwnerReferences[0].Name != "TestService" {
		t.Fatalf("Ownerreference Name does not match: %v", actual.ObjectMeta.OwnerReferences[0].Name)
	}
}

func ExpectServiceOwnerReference(actual *corev1.Service, instance *esov1alpha1.ExternalService, t *testing.T) {
	if len(actual.ObjectMeta.OwnerReferences) != 1 {
		t.Fatalf("Ownerreference missing: %v", actual)
	}

	if actual.ObjectMeta.OwnerReferences[0].Kind != "ExternalService" {
		t.Fatalf("Ownerreference Kind does not match: %v", actual.ObjectMeta.OwnerReferences[0].Kind)
	}

	if actual.ObjectMeta.OwnerReferences[0].APIVersion != "eso.crowdfox.com/v1alpha1" {
		t.Fatalf("Ownerreference API does not match: %v", actual.ObjectMeta.OwnerReferences[0].APIVersion)
	}

	if actual.ObjectMeta.OwnerReferences[0].Name != "TestService" {
		t.Fatalf("Ownerreference Name does not match: %v", actual.ObjectMeta.OwnerReferences[0].Name)
	}
}

func ExpectIngressOwnerReference(actual *extv1.Ingress, instance *esov1alpha1.ExternalService, t *testing.T) {
	if len(actual.ObjectMeta.OwnerReferences) != 1 {
		t.Fatalf("Ownerreference missing: %v", actual)
	}

	if actual.ObjectMeta.OwnerReferences[0].Kind != "ExternalService" {
		t.Fatalf("Ownerreference Kind does not match: %v", actual.ObjectMeta.OwnerReferences[0].Kind)
	}

	if actual.ObjectMeta.OwnerReferences[0].APIVersion != "eso.crowdfox.com/v1alpha1" {
		t.Fatalf("Ownerreference API does not match: %v", actual.ObjectMeta.OwnerReferences[0].APIVersion)
	}

	if actual.ObjectMeta.OwnerReferences[0].Name != "TestService" {
		t.Fatalf("Ownerreference Name does not match: %v", actual.ObjectMeta.OwnerReferences[0].Name)
	}
}

func ExpectEndpointWithNameAndNamespace(actual *corev1.Endpoints, name string, namespace string, t *testing.T) {
	if actual.Name != name {
		t.Errorf("Created Endpoint has Name: '%v' but '%v' was expected", actual.Name, name)
	}

	if actual.Namespace != namespace {
		t.Errorf("Created Endpoint has Name: '%v' but '%v' was expected", actual.Namespace, namespace)
	}
}

func ExpectServiceWithNameAndNamespace(actual *corev1.Service, name string, namespace string, t *testing.T) {
	if actual.Name != name {
		t.Errorf("Created Service has Name: '%v' but '%v' was expected", actual.Name, name)
	}

	if actual.Namespace != namespace {
		t.Errorf("Created Service has Name: '%v' but '%v' was expected", actual.Namespace, namespace)
	}
}

func ExpectIngressWithNameAndNamespace(actual *extv1.Ingress, name string, namespace string, t *testing.T) {
	if actual.Name != name {
		t.Errorf("Created Ingress has Name: '%v' but '%v' was expected", actual.Name, name)
	}

	if actual.Namespace != namespace {
		t.Errorf("Created Ingress has Name: '%v' but '%v' was expected", actual.Namespace, namespace)
	}
}

func ExpectEqualEndpoints(actual, expected corev1.Endpoints, t *testing.T) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected Endpoint: %v  to have the same values as Endpoint: %v but it has not:\n  (actual)  %v\n!=\n  (expected)%v\n", actual.Name, expected.Name, actual, expected)
	}
}

func ExpectNoErrorsAndRequeue(res reconcile.Result, err error, t *testing.T) {
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	if res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
}
