package externalservice

import (
	"github.com/CrowdfoxGmbH/external-service-operator/pkg/testutils"
	"testing"
)

func TestCreateIngressCr(t *testing.T) {
	ingress := createIngressCr(getTestExternalServiceCR())

	testutils.ExpectEqStr(ingress.Name, "TestService", t)
	testutils.ExpectEqStr(ingress.Namespace, "external-services", t)

	testutils.ExpectEqStr(ingress.Annotations["foo.bar"], "testvalue", t)

	testutils.ExpectEqStr(ingress.Spec.Rules[0].Host, "subdomain.example.com", t)
	testutils.ExpectEqStr(ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path, "", t)
	testutils.ExpectEqStr(ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName, "TestService", t)
	testutils.ExpectEqStr(ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServicePort.String(), "80", t)

	testutils.ExpectEqStr(ingress.Spec.Rules[1].Host, "another.domain.com", t)
	testutils.ExpectEqStr(ingress.Spec.Rules[1].IngressRuleValue.HTTP.Paths[0].Path, "/foo", t)

	testutils.ExpectEqStr(ingress.Spec.Rules[1].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName, "TestService", t)

	testutils.ExpectEqStr(ingress.Spec.Rules[1].IngressRuleValue.HTTP.Paths[0].Backend.ServicePort.String(), "80", t)

}
