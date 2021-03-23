package prober

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/CrowdfoxGmbH/external-service-operator/pkg/testutils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/probe"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/go-logr/logr"
	"testing"
)

func TestEnsureUnreadyOfUnReadyIp(t *testing.T) {
	endpoint := testutils.CreateDefaultEndpoint()

	client := fake.NewFakeClient(endpoint)
	worker := worker{
		client: client,
		ip:     "10.0.102.16",
	}

	if err := worker.ensureUnready(endpoint); err != nil {
		t.Errorf("Got error '%v' while executing function under test", err)
	}

	actualEndpoint := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, actualEndpoint); err != nil {
		t.Errorf("Got Error '%v' getting updated Endpoint", err)
	}

	if actualEndpoint.Name == "" {
		t.Errorf("Got empty Name for Endpoint")
	}
	if actualEndpoint.Subsets[0].NotReadyAddresses[0].IP != "10.0.102.14" {
		t.Errorf("First NotReadyIP is not anymore on the first spot")
	}
	if actualEndpoint.Subsets[0].NotReadyAddresses[1].IP != "10.0.102.16" {
		t.Errorf("Second NotReady IP is not in the unready list anymore")
	}
	if actualEndpoint.Subsets[0].Addresses[0].IP != "10.0.102.10" {
		t.Errorf("First ReadyIP is not anymore on the first spot")
	}
	if actualEndpoint.Subsets[0].Addresses[1].IP != "10.0.102.12" {
		t.Errorf("Second ReadyIP is not anymore on the second spot")
	}
}

func TestEnsureUnreadyOfReadyIp(t *testing.T) {
	endpoint := testutils.CreateDefaultEndpoint()

	client := fake.NewFakeClient(endpoint)
	worker := worker{
		client: client,
		ip:     "10.0.102.10",
	}

	if err := worker.ensureUnready(endpoint); err != nil {
		t.Errorf("Got error '%v' while executing function under test", err)
	}

	actualEndpoint := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, actualEndpoint); err != nil {
		t.Errorf("Got Error '%v' getting updated Endpoint", err)
	}

	if actualEndpoint.Name == "" {
		t.Errorf("Got empty Name for Endpoint")
	}
	if actualEndpoint.Subsets[0].NotReadyAddresses[0].IP != "10.0.102.14" {
		t.Errorf("First NotReadyIP is not anymore on the first spot")
	}
	if actualEndpoint.Subsets[0].NotReadyAddresses[1].IP != "10.0.102.16" {
		t.Errorf("Second NotReady IP is not on the correct spot anymore")
	}
	if actualEndpoint.Subsets[0].NotReadyAddresses[2].IP != "10.0.102.10" {
		t.Errorf("New IP was not added to NotReadyAddresses")
	}
	if actualEndpoint.Subsets[0].Addresses[0].IP != "10.0.102.12" {
		t.Errorf("Expected IP not to be touched but was removed")
	}
}

func TestEnsureReadyOfReadyIp(t *testing.T) {
	endpoint := testutils.CreateDefaultEndpoint()

	client := fake.NewFakeClient(endpoint)
	worker := worker{
		client: client,
		ip:     "10.0.102.10",
	}

	if err := worker.ensureReady(endpoint); err != nil {
		t.Errorf("Got error '%v' while executing function under test", err)
	}

	actualEndpoint := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, actualEndpoint); err != nil {
		t.Errorf("Got Error '%v' getting updated Endpoint", err)
	}

	if actualEndpoint.Name == "" {
		t.Errorf("Got empty Name for Endpoint")
	}
	if actualEndpoint.Subsets[0].NotReadyAddresses[0].IP != "10.0.102.14" {
		t.Errorf("First Unready IP is not anymore on the first spot")
	}
	if actualEndpoint.Subsets[0].NotReadyAddresses[1].IP != "10.0.102.16" {
		t.Errorf("Second Unready IP is not on the correct spot anymore")
	}
	if actualEndpoint.Subsets[0].Addresses[0].IP != "10.0.102.10" {
		t.Errorf("First Ready IP was not on the correct spot anymore")
	}
	if actualEndpoint.Subsets[0].Addresses[1].IP != "10.0.102.12" {
		t.Errorf("Second Ready IP was not on the correct spot anymore")
	}
}

func TestEnsureReadyOfUnReadyIp(t *testing.T) {
	endpoint := testutils.CreateDefaultEndpoint()

	client := fake.NewFakeClient(endpoint)
	worker := worker{
		client: client,
		ip:     "10.0.102.14",
	}

	if err := worker.ensureReady(endpoint); err != nil {
		t.Errorf("Got error '%v' while executing function under test", err)
	}

	actualEndpoint := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, actualEndpoint); err != nil {
		t.Errorf("Got Error '%v' getting updated Endpoint", err)
	}

	if actualEndpoint.Name == "" {
		t.Errorf("Got empty Name for Endpoint")
	}
	if actualEndpoint.Subsets[0].NotReadyAddresses[0].IP != "10.0.102.16" {
		t.Errorf("First Unready IP should be the only one left")
	}
	if actualEndpoint.Subsets[0].Addresses[0].IP != "10.0.102.10" {
		t.Errorf("First Ready IP was not on the correct spot anymore")
	}
	if actualEndpoint.Subsets[0].Addresses[1].IP != "10.0.102.12" {
		t.Errorf("Second Ready IP was not on the correct spot anymore")
	}
	if actualEndpoint.Subsets[0].Addresses[2].IP != "10.0.102.14" {
		t.Errorf("IP was not added to the Ready Addresses")
	}
}

func TestDoProbeEndpointDisappeared(t *testing.T) {
	fakeLogger := testLogger{}
	log = &fakeLogger
	client := fake.NewFakeClient() //given no endpoint ressource
	worker := worker{
		namespacedName: types.NamespacedName{Name: "TestService", Namespace: "external-services"},
		client:         client,
		ip:             "Fake",
		probe:          testutils.CreateTestProbe(1, 5, 3, 3, 3, corev1.URISchemeHTTP, 80, "/"),
	}

	keepGoing := worker.doProbe()

	// expect worker will stop
	if keepGoing {
		t.Errorf("Expected to stop Healthcheck worker when endpoint could not be found")
	}

	// and log Endpoint could not be found
	fakeLogger.expectErrorLog("Endpoint could not be found", t)
}

func TestDoProbeWithoutAnyProbeType(t *testing.T) {
	fakeLogger := testLogger{}
	log = &fakeLogger

	endpoint := testutils.CreateDefaultEndpoint()
	client := fake.NewFakeClient(endpoint)
	probeWithoutHandler := testutils.CreateDefaultTestProbe()
	probeWithoutHandler.Handler = corev1.Handler{}

	worker := worker{
		parent: &externalServiceProber{
			httpprober: newFakeHTTPProber(Error),
			tcpprober:  newFakeTCPProber(Error),
		},
		namespacedName: types.NamespacedName{Name: "TestService", Namespace: "external-services"},
		client:         client,
		ip:             "127.0.0.1",
		probe:          probeWithoutHandler,
	}

	keepGoing := worker.doProbe()

	// expect worker will stop
	if keepGoing {
		t.Errorf("Expected to stop Healthcheck worker when healthcheck has failures")
	}

	// and log Endpoint could not be found
	fakeLogger.expectErrorLog("Unknown Probe Type.", t)

}

func TestTcpProbeSuccessfulOfInactiveService(t *testing.T) {
	fakeLogger := testLogger{}
	log = &fakeLogger

	endpoint := testutils.CreateDefaultEndpoint()
	client := fake.NewFakeClient(endpoint)
	tcpProbe := testutils.CreateDefaultTestProbe()
	tcpProbe.Handler = corev1.Handler{
		TCPSocket: &corev1.TCPSocketAction{
			Port: intstr.FromInt(80),
			Host: "not used",
		},
	}
	tcpProbe.SuccessThreshold = 1

	worker := worker{
		parent: &externalServiceProber{
			httpprober: newFakeHTTPProber(Error),
			tcpprober:  newFakeTCPProber(Success),
		},
		namespacedName: types.NamespacedName{Name: "TestService", Namespace: "external-services"},
		client:         client,
		ip:             "10.0.102.14",
		probe:          tcpProbe,
	}

	keepGoing := worker.doProbe()

	if !keepGoing {
		t.Errorf("Error during Healthcheck even though TCP Check was successfull")
	}

	actualEndpoint := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, actualEndpoint); err != nil {
		t.Errorf("Got Error '%v' getting updated Endpoint", err)
	}

	// 10.0.102.10 should not be active anymore
	if actualEndpoint.Subsets[0].NotReadyAddresses[0].IP == "10.0.102.14" {
		t.Errorf("Endpoint is expected to healthy, but still in unhealthy list")
	}

	if actualEndpoint.Subsets[0].Addresses[2].IP != "10.0.102.14" {
		t.Errorf("Endpoint is expected to healthy, but not in healthy list")
	}

}

func TestTcpProbeFailureOfActiveService(t *testing.T) {
	fakeLogger := testLogger{}
	log = &fakeLogger

	endpoint := testutils.CreateDefaultEndpoint()
	client := fake.NewFakeClient(endpoint)
	tcpProbe := testutils.CreateDefaultTestProbe()
	tcpProbe.Handler = corev1.Handler{
		TCPSocket: &corev1.TCPSocketAction{
			Port: intstr.FromInt(80),
			Host: "not used",
		},
	}
	tcpProbe.FailureThreshold = 1

	worker := worker{
		parent: &externalServiceProber{
			httpprober: newFakeHTTPProber(Error),
			tcpprober:  newFakeTCPProber(Failure),
		},
		namespacedName: types.NamespacedName{Name: "TestService", Namespace: "external-services"},
		client:         client,
		ip:             "10.0.102.10",
		probe:          tcpProbe,
	}

	keepGoing := worker.doProbe()

	if !keepGoing {
		t.Errorf("Unexpected error during Healthcheck")
	}

	actualEndpoint := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, actualEndpoint); err != nil {
		t.Errorf("Got Error '%v' getting updated Endpoint", err)
	}

	// 10.0.102.10 should not be active anymore
	if actualEndpoint.Subsets[0].Addresses[0].IP == "10.0.102.10" {
		t.Errorf("Endpoint is expected to be unhealthy, but still there")
	}

}

func TestDoProbeWithInvalidIp(t *testing.T) {
	fakeLogger := testLogger{}
	log = &fakeLogger

	endpoint := testutils.CreateDefaultEndpoint()
	client := fake.NewFakeClient(endpoint)

	worker := worker{
		parent: &externalServiceProber{
			httpprober: newFakeHTTPProber(Error),
		},
		namespacedName: types.NamespacedName{Name: "TestService", Namespace: "external-services"},
		client:         client,
		ip:             "Fake",
		probe:          testutils.CreateTestProbe(1, 5, 3, 3, 3, corev1.URISchemeHTTP, 80, "/"),
	}

	keepGoing := worker.doProbe()

	// expect worker will stop
	if keepGoing {
		t.Errorf("Expected to stop Healthcheck worker when healthcheck has failures")
	}

	// and log Endpoint could not be found
	fakeLogger.expectErrorLog("Runtimeerror during probe", t)
}

func TestDoProbeStillHealthy(t *testing.T) {
	fakeLogger := testLogger{}
	log = &fakeLogger

	endpoint := testutils.CreateDefaultEndpoint()
	client := fake.NewFakeClient(endpoint)

	worker := worker{
		parent: &externalServiceProber{
			httpprober: newFakeHTTPProber(Success),
		},
		namespacedName: types.NamespacedName{Name: "TestService", Namespace: "external-services"},
		client:         client,
		ip:             "10.0.102.10",
		probe:          testutils.CreateTestProbe(1, 5, 3, 3, 1, corev1.URISchemeHTTP, 80, "/"),
	}

	keepGoing := worker.doProbe()

	if !keepGoing {
		t.Errorf("Expected to stop Healthcheck worker when healthcheck has failures")
	}

	actualEndpoint := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, actualEndpoint); err != nil {
		t.Errorf("Got Error '%v' getting updated Endpoint", err)
	}

	fakeLogger.expectNotInfoLog("Update Endpoint, because availability changed", t)

}

func TestDoProbeFailedOnceAllowedThree(t *testing.T) {
	fakeLogger := testLogger{}
	log = &fakeLogger

	endpoint := testutils.CreateDefaultEndpoint()
	client := fake.NewFakeClient(endpoint)

	failureThreshold := int32(3)

	worker := worker{
		parent: &externalServiceProber{
			httpprober: newFakeHTTPProber(Failure),
		},
		namespacedName: types.NamespacedName{Name: "TestService", Namespace: "external-services"},
		client:         client,
		ip:             "10.0.102.10",
		probe:          testutils.CreateTestProbe(1, 5, 3, 3, failureThreshold, corev1.URISchemeHTTP, 80, "/"),
	}

	keepGoing := worker.doProbe()

	if !keepGoing {
		t.Errorf("Expected to stop Healthcheck worker when healthcheck has failures")
	}

	actualEndpoint := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, actualEndpoint); err != nil {
		t.Errorf("Got Error '%v' getting updated Endpoint", err)
	}

	if actualEndpoint.Subsets[0].Addresses[0].IP != "10.0.102.10" {
		t.Errorf("Endpoint is expected to be healthy, but was (re)moved")
	}

}

func TestDoProbeFailedOnceAllowedNone(t *testing.T) {
	fakeLogger := testLogger{}
	log = &fakeLogger

	endpoint := testutils.CreateDefaultEndpoint()
	client := fake.NewFakeClient(endpoint)

	failureThreshold := int32(1)

	worker := worker{
		parent: &externalServiceProber{
			httpprober: newFakeHTTPProber(Failure),
		},
		namespacedName: types.NamespacedName{Name: "TestService", Namespace: "external-services"},
		client:         client,
		ip:             "10.0.102.10",
		probe:          testutils.CreateTestProbe(1, 5, 3, 3, failureThreshold, corev1.URISchemeHTTP, 80, "/"),
	}

	keepGoing := worker.doProbe()
	if !keepGoing {
		t.Errorf("Expected to stop Healthcheck worker when healthcheck has failures")
	}

	actualEndpoint := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, actualEndpoint); err != nil {
		t.Errorf("Got Error '%v' getting updated Endpoint", err)
	}

	if actualEndpoint.Subsets[0].NotReadyAddresses[2].IP != "10.0.102.10" {
		t.Errorf("Endpoint is expected to be unhealthy, but was not moved")
	}

}

func TestDoProbeFailedThreeTimesAllowedTwo(t *testing.T) {
	fakeLogger := testLogger{}
	log = &fakeLogger

	endpoint := testutils.CreateDefaultEndpoint()
	client := fake.NewFakeClient(endpoint)

	failureThreshold := int32(3)

	worker := worker{
		parent: &externalServiceProber{
			httpprober: newFakeHTTPProber(Failure),
		},
		namespacedName: types.NamespacedName{Name: "TestService", Namespace: "external-services"},
		client:         client,
		ip:             "10.0.102.10",
		probe:          testutils.CreateTestProbe(1, 5, 3, 3, failureThreshold, corev1.URISchemeHTTP, 80, "/"),
	}

	worker.doProbe()
	worker.doProbe()
	keepGoing := worker.doProbe()

	if !keepGoing {
		t.Errorf("Expected to stop Healthcheck worker when healthcheck has failures")
	}

	actualEndpoint := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, actualEndpoint); err != nil {
		t.Errorf("Got Error '%v' getting updated Endpoint", err)
	}

	if actualEndpoint.Subsets[0].NotReadyAddresses[2].IP != "10.0.102.10" {
		t.Errorf("Endpoint is expected to be unhealthy, but was not moved")
	}

}

func TestDoProbeFailed2TimesSuccessfullOnce2TimeFailure(t *testing.T) {
	fakeLogger := testLogger{}
	log = &fakeLogger

	endpoint := testutils.CreateDefaultEndpoint()
	client := fake.NewFakeClient(endpoint)

	failureThreshold := int32(3)

	worker := worker{
		parent: &externalServiceProber{
			httpprober: newFakeHTTPProber(Failure),
		},
		namespacedName: types.NamespacedName{Name: "TestService", Namespace: "external-services"},
		client:         client,
		ip:             "10.0.102.10",
		probe:          testutils.CreateTestProbe(1, 5, 3, 3, failureThreshold, corev1.URISchemeHTTP, 80, "/"),
	}

	worker.doProbe()
	worker.doProbe()
	worker.parent.httpprober = newFakeHTTPProber(Success)
	keepGoing := worker.doProbe()

	if !keepGoing {
		t.Errorf("Expected to stop Healthcheck worker when healthcheck has failures")
	}

	actualEndpoint := &corev1.Endpoints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: endpoint.Name, Namespace: endpoint.Namespace}, actualEndpoint); err != nil {
		t.Errorf("Got Error '%v' getting updated Endpoint", err)
	}

	if actualEndpoint.Subsets[0].Addresses[0].IP != "10.0.102.10" {
		t.Errorf("Endpoint is expected to be healthy, but was (re)moved")
	}
}

func TestRunHttpProbe(t *testing.T) {
	probe := testutils.CreateDefaultTestProbe()
	httpaction := probe.HTTPGet
	httpaction.Host = "really.needed.domain"
	actualHeaders := buildHeader(httpaction)

	if actualHeaders.Get("Host") != "really.needed.domain" {
		t.Errorf("Expected Host header to be added when Host value is set on HTTPGetAction in Probe")
	}
}

type testLogger struct {
	errorLogs []string
	infoLogs  []string
}

func (t *testLogger) expectErrorLog(msg string, test *testing.T) bool {
	for _, message := range t.errorLogs {
		if message == msg {
			return true
		}
	}

	test.Errorf("Could not found msg: %v in messages %v", msg, t.errorLogs)
	return false
}

func (t *testLogger) expectNotInfoLog(msg string, test *testing.T) bool {
	for _, message := range t.infoLogs {
		if message == msg {
			test.Errorf("Found unexpected msg: %v in info log messages", msg)
			return false
		}
	}
	return true

}

func (t *testLogger) expectInfoLog(msg string, test *testing.T) bool {
	for _, message := range t.infoLogs {
		if message == msg {
			return true
		}
	}
	test.Errorf("Could not found msg: %v in messages %v", msg, t.infoLogs)
	return false
}

func (t *testLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	t.errorLogs = append(t.errorLogs, msg)
}

func (t *testLogger) Info(msg string, keysAndValues ...interface{}) {
	t.infoLogs = append(t.infoLogs, msg)
}

func (t *testLogger) V(level int) logr.InfoLogger {
	return t
}

func (t *testLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return t
}

func (t *testLogger) WithName(name string) logr.Logger {
	return t
}

func (t *testLogger) Enabled() bool {
	return true
}

type fakeHTTPAnswer string

const (
	Error   fakeHTTPAnswer = "error"
	Failure fakeHTTPAnswer = "failure"
	Unknown fakeHTTPAnswer = "unknown"
	Success fakeHTTPAnswer = "success"
)

type fakeHTTPProber struct {
	answer fakeHTTPAnswer
}

func newFakeHTTPProber(answer fakeHTTPAnswer) *fakeHTTPProber {
	return &fakeHTTPProber{
		answer: answer,
	}
}

func (p *fakeHTTPProber) Probe(_ *url.URL, _ http.Header, _ time.Duration) (probe.Result, string, error) {
	switch p.answer {
	case Error:
		return probe.Failure, "Fake error", errors.New("Error")
	case Failure:
		return probe.Failure, "Fake error", nil
	case Success:
		return probe.Success, "Success", nil
	case Unknown:
		return probe.Unknown, "Fake error", errors.New("Error")
	}

	return probe.Failure, "Not Implemented", errors.New("Fake is broken")
}

type fakeTCPProber struct {
	answer fakeHTTPAnswer
}

func newFakeTCPProber(answer fakeHTTPAnswer) *fakeTCPProber {
	return &fakeTCPProber{
		answer: answer,
	}
}

func (p *fakeTCPProber) Probe(host string, port int, timeout time.Duration) (probe.Result, string, error) {
	switch p.answer {
	case Error:
		return probe.Failure, "Fake error", errors.New("Error")
	case Failure:
		return probe.Failure, "Fake error", nil
	case Success:
		return probe.Success, "Success", nil
	case Unknown:
		return probe.Unknown, "Fake error", errors.New("Error")
	}

	return probe.Failure, "Not Implemented", errors.New("Fake is broken")
}
