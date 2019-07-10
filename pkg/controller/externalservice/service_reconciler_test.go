package externalservice

import (
	"github.com/CrowdfoxGmbH/external-service-operator/pkg/testutils"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestCreateServiceCr(t *testing.T) {

	service := createServiceCr(getTestExternalServiceCR())

	//check name is set
	testutils.ExpectEqStr(service.Name, "TestService", t)
	testutils.ExpectEqStr(service.Namespace, "external-services", t)

	testutils.ExpectEqInt(service.Spec.Ports[0].Port, 80, t)

	testutils.ExpectEqStr(service.Spec.ClusterIP, "None", t)

	testutils.ExpectServiceType(service.Spec.Type, corev1.ServiceTypeClusterIP, t)

}
