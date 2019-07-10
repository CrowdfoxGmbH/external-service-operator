package prober

import (
	"context"

	"github.com/CrowdfoxGmbH/external-service-operator/pkg/testutils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"testing"
)

func TestAddProbes(t *testing.T) {
	externalService := testutils.CreateDefaultExternalService()
	externalService.Spec.ReadinessProbe = testutils.CreateDefaultTestProbe()

	client := testutils.InitFakeClient(externalService)
	prober := NewProber(client)
	if prober == nil {
		t.Errorf("Prober must not be nil")
	}

	prober.AddProbes(externalService)

	if probe, ok := prober.probes[types.NamespacedName{Name: externalService.Name, Namespace: externalService.Namespace}]; !ok {
		t.Errorf("Could not find expected EndpointProbe")
	} else {
		if count := len(probe.workers); count != 3 {
			t.Errorf("Expected 4 workers, but got %v", count)
		}
	}
}

func TestAddProbesForEndpointWithoutProbe(t *testing.T) {
	externalService := testutils.CreateDefaultExternalService()
	endpoint := testutils.CreateDefaultEndpoint()

	client := testutils.InitFakeClient(externalService, endpoint)
	prober := NewProber(client)

	prober.AddProbes(externalService)

	found := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, found); err != nil {
		t.Errorf("Endpoint does not exist")
		return
	}

	// expect all IPs should be in Ready Addresses
	testutils.ExpectEqStr(found.Subsets[0].Addresses[0].IP, "10.0.102.10", t)
	testutils.ExpectEqStr(found.Subsets[0].Addresses[1].IP, "10.0.102.12", t)
	testutils.ExpectEqStr(found.Subsets[0].Addresses[2].IP, "10.0.102.14", t)
	testutils.ExpectEqStr(found.Subsets[0].Addresses[3].IP, "10.0.102.16", t)
}

func TestRemoveProbes(t *testing.T) {
	externalService := testutils.CreateDefaultExternalService()
	externalService.Spec.ReadinessProbe = testutils.CreateDefaultTestProbe()

	otherExternalService := testutils.CreateExternalService("OtherService", "external-services", []string{"10.0.102.10", "10.0.102.12", "10.0.102.14"}, 9191, testutils.CreateDefaultExternalHostPaths())
	otherExternalService.Spec.ReadinessProbe = testutils.CreateDefaultTestProbe()

	client := testutils.InitFakeClient(externalService)
	prober := NewProber(client)
	if prober == nil {
		t.Errorf("Prober must not be nil")
	}

	prober.AddProbes(externalService)
	prober.AddProbes(otherExternalService)

	prober.RemoveProbes(externalService)

	if _, found := prober.probes[types.NamespacedName{Name: externalService.Name, Namespace: externalService.Namespace}]; found {
		t.Errorf("Expected ExternalService Probes to be deleted for %v", externalService.Name)
	}

	if _, found := prober.probes[types.NamespacedName{Name: otherExternalService.Name, Namespace: otherExternalService.Namespace}]; !found {
		t.Errorf("Expected Other Service not to be deleted")
	}
}
