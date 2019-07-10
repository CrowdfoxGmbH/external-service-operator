package externalservice

import (
	esov1alpha1 "github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/CrowdfoxGmbH/external-service-operator/pkg/testutils"
	"k8s.io/apimachinery/pkg/api/errors"
	"testing"
)

func getTestExternalServiceCR() *esov1alpha1.ExternalService {
	return &esov1alpha1.ExternalService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestService",
			Namespace: "external-services",
			Annotations: map[string]string{
				"foo.bar": "testvalue",
			},
		},

		Spec: esov1alpha1.ExternalServiceSpec{
			Port: 80,
			Ips: []string{
				"10.0.100.10",
				"10.0.100.11",
				"10.0.100.12",
			},
			Hosts: []esov1alpha1.ExternalServiceHostPath{
				esov1alpha1.ExternalServiceHostPath{
					Host: "subdomain.example.com",
				},
				esov1alpha1.ExternalServiceHostPath{
					Host: "another.domain.com",
					Path: "/foo",
				},
			},
		},
	}
}

func TestReconcileCreateEndpoints(t *testing.T) {
	instance := getTestExternalServiceCR()

	cl := testutils.InitFakeClient(instance)

	res, err := runTestReconcile(cl, instance.Name, instance.Namespace)

	testutils.ExpectNoErrorsAndRequeue(res, err, t)

	actualEndpoint, err := getRuntimeEndpoint(cl, instance.Name, instance.Namespace)
	if err != nil {
		t.Fatalf("get Endpoint: (%v)", err)
	}

	testutils.ExpectEndpointWithNameAndNamespace(actualEndpoint, "TestService", "external-services", t)
	testutils.ExpectEndpointOwnerReference(actualEndpoint, instance, t)
}

func TestUpdateIpsOnEndpoints(t *testing.T) {
	// Given
	oldInstance := getTestExternalServiceCR()
	oldEndpoint := CreateEndpointsCr(oldInstance)
	oldEndpoint.ObjectMeta.ResourceVersion = "4711"

	client := testutils.InitFakeClient(oldInstance, oldEndpoint)

	// When
	//  I change the IP
	newInstance := oldInstance.DeepCopy()
	newInstance.Spec.Ips = []string{
		"123.456.789",
		"8.8.8.8",
	}
	newInstance.Spec.ReadinessProbe = testutils.CreateDefaultTestProbe() //without probes all IPs would be moved to ready
	updateObject(client, newInstance)
	res, err := runTestReconcile(client, newInstance.Name, newInstance.Namespace)

	// Then
	//  it should successfully update the IPs on the Endpoint without errors
	testutils.ExpectNoErrorsAndRequeue(res, err, t)
	actualEndpoint, err := getRuntimeEndpoint(client, newInstance.Name, oldInstance.Namespace)
	if err != nil {
		t.Fatalf("get Endpoint: (%v)", err)
	}

	testutils.ExpectEqStr(actualEndpoint.Subsets[0].NotReadyAddresses[0].IP, "123.456.789", t)
	testutils.ExpectEqStr(actualEndpoint.Subsets[0].NotReadyAddresses[1].IP, "8.8.8.8", t)
	testutils.ExpectEqStr(actualEndpoint.ResourceVersion, "4711", t)
}

func TestReconcileService(t *testing.T) {
	instance := getTestExternalServiceCR()
	client := testutils.InitFakeClient(instance)

	res, err := runTestReconcile(client, instance.Name, instance.Namespace)
	testutils.ExpectNoErrorsAndRequeue(res, err, t)

	actualService, err := getRuntimeService(client, instance.Name, instance.Namespace)
	if err != nil {
		t.Fatalf("get Service: (%v)", err)
	}

	testutils.ExpectServiceWithNameAndNamespace(actualService, "TestService", "external-services", t)
	testutils.ExpectServiceOwnerReference(actualService, instance, t)
}

func TestReconcileServiceUpdatePort(t *testing.T) {
	// Given an old ExternalService CR applied
	oldInstance := getTestExternalServiceCR()
	oldService := createServiceCr(oldInstance)
	oldService.ResourceVersion = "1337"
	client := testutils.InitFakeClient(oldInstance, oldService)

	// When I update the Port Attribute on my ExternalService CR
	newInstance := oldInstance.DeepCopy()
	newInstance.Spec.Port = 8080
	updateObject(client, newInstance)

	res, err := runTestReconcile(client, newInstance.Name, newInstance.Namespace)
	testutils.ExpectNoErrorsAndRequeue(res, err, t)

	// Then the Port should be 8080 on my Service
	actualService, err := getRuntimeService(client, newInstance.Name, newInstance.Namespace)
	if err != nil {
		t.Fatalf("get Service: (%v)", err)
	}

	testutils.ExpectEqInt(actualService.Spec.Ports[0].Port, 8080, t)
	// check if the same version is provided. This is used as optimistic locking
	testutils.ExpectEqStr(actualService.ObjectMeta.ResourceVersion, "1337", t)
	testutils.ExpectEqInt(int32(len(actualService.ObjectMeta.OwnerReferences)), 1, t)
}

func TestReconcileIngress(t *testing.T) {
	instance := getTestExternalServiceCR()
	client := testutils.InitFakeClient(instance)

	res, err := runTestReconcile(client, instance.Name, instance.Namespace)
	testutils.ExpectNoErrorsAndRequeue(res, err, t)

	actualIngress, err := getRuntimeIngress(client, instance.Name, instance.Namespace)
	if err != nil {
		t.Fatalf("get Service: (%v)", err)
	}

	testutils.ExpectIngressWithNameAndNamespace(actualIngress, "TestService", "external-services", t)
	testutils.ExpectIngressOwnerReference(actualIngress, instance, t)
}

func TestReconcileIngressWithoutHost(t *testing.T) {
	instance := getTestExternalServiceCR()
	instance.Spec.Hosts = []esov1alpha1.ExternalServiceHostPath{}

	client := testutils.InitFakeClient(instance)

	res, err := runTestReconcile(client, instance.Name, instance.Namespace)
	testutils.ExpectNoErrorsAndRequeue(res, err, t)

	found, err := getRuntimeIngress(client, instance.Name, instance.Namespace)
	if err == nil || !errors.IsNotFound(err) {
		t.Fatalf("Expected that no Ingress was created, but found %v", found)
	}

}

func TestReconcileIngressDeleteRSWhenHostsAreRemoved(t *testing.T) {
	instance := getTestExternalServiceCR()
	oldIngress := createIngressCr(instance.DeepCopy())
	instance.Spec.Hosts = []esov1alpha1.ExternalServiceHostPath{}

	client := testutils.InitFakeClient(instance, oldIngress)

	res, err := runTestReconcile(client, instance.Name, instance.Namespace)
	testutils.ExpectNoErrorsAndRequeue(res, err, t)

	found, err := getRuntimeIngress(client, instance.Name, instance.Namespace)
	if err == nil || !errors.IsNotFound(err) {
		t.Fatalf("Expected that Ingress was deleted, but found %v", found)
	}

}

func TestReconcileIngressAddHostpathAndRemoveAnnotation(t *testing.T) {
	// Given an old ExternalService CR applied
	oldInstance := getTestExternalServiceCR()
	oldIngress := createIngressCr(oldInstance)
	oldIngress.ResourceVersion = "2019"
	client := testutils.InitFakeClient(oldInstance, oldIngress)

	// When I add a new subpath
	newInstance := oldInstance.DeepCopy()
	newInstance.Spec.Hosts[0].Path = "/newSubpath"
	// And remove the annotation
	newInstance.ObjectMeta.Annotations = map[string]string{}
	updateObject(client, newInstance)

	res, err := runTestReconcile(client, newInstance.Name, newInstance.Namespace)
	testutils.ExpectNoErrorsAndRequeue(res, err, t)

	// Then the Port should have no annotations and a new subpath
	actualIngress, err := getRuntimeIngress(client, newInstance.Name, newInstance.Namespace)
	if err != nil {
		t.Fatalf("get Service: (%v)", err)
	}

	testutils.ExpectTrue(len(actualIngress.ObjectMeta.Annotations) == 0, t)
	testutils.ExpectEqStr(actualIngress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path, "/newSubpath", t)
	// check if the same version is provided. This is used as optimistic locking
	testutils.ExpectEqStr(actualIngress.ResourceVersion, "2019", t)
	testutils.ExpectEqInt(int32(len(actualIngress.ObjectMeta.OwnerReferences)), 1, t)
}
