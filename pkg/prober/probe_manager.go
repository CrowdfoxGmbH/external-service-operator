package prober

import (
	"context"

	esov1alpha1 "github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	httpprober "k8s.io/kubernetes/pkg/probe/http"
	tcpprober "k8s.io/kubernetes/pkg/probe/tcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type ProbeManager struct {
	client client.Client
	probes map[types.NamespacedName]*externalServiceProber
	logger logr.Logger
}

func NewProber(client client.Client) *ProbeManager {
	return &ProbeManager{
		client: client,
		probes: map[types.NamespacedName]*externalServiceProber{},
		logger: logf.Log.WithName("Probe Manager"),
	}
}

func (p *ProbeManager) markAddressesReady(externalService *esov1alpha1.ExternalService) {
	found := &corev1.Endpoints{}
	key := types.NamespacedName{Name: externalService.Name, Namespace: externalService.Namespace}
	if err := p.client.Get(context.TODO(), key, found); err != nil {
		p.logger.Error(err, "Could not get Endpoint")
		return
	}

	for _, notReady := range found.Subsets[0].NotReadyAddresses {
		found.Subsets[0].Addresses = append(found.Subsets[0].Addresses, notReady)
	}
	found.Subsets[0].NotReadyAddresses = []corev1.EndpointAddress{}

	if err := p.client.Update(context.TODO(), found); err != nil {
		p.logger.Error(err, "Could update Endpoint")
		return
	}
}

func (p *ProbeManager) AddProbes(externalService *esov1alpha1.ExternalService) {
	if externalService.Spec.ReadinessProbe == (corev1.Probe{}) {
		p.logger.Info("External Service does not have a Probe. Mark all Addresses ready.", "externalservice", externalService.Name)
		p.markAddressesReady(externalService)
		return
	}

	key := types.NamespacedName{Name: externalService.Name, Namespace: externalService.Namespace}

	if _, found := p.probes[key]; found {
		return
	}

	prober := &externalServiceProber{
		externalService: externalService,
		workers:         map[string]*worker{},
		httpprober:      httpprober.New(),
		tcpprober:       tcpprober.New(),
	}

	prober.addWorkers(p.client, externalService.Spec.ReadinessProbe)
	p.probes[key] = prober
}

func (p *ProbeManager) RemoveProbes(externalService *esov1alpha1.ExternalService) {
	key := types.NamespacedName{Name: externalService.Name, Namespace: externalService.Namespace}
	p.RemoveProbesByNamespacedName(key)
}

func (p *ProbeManager) RemoveProbesByNamespacedName(key types.NamespacedName) {
	if probe, found := p.probes[key]; found {
		probe.shutdownAllWorkers()
		delete(p.probes, key)
	}
}

func (p *ProbeManager) UpdateProbes(externalService *esov1alpha1.ExternalService) {
	key := types.NamespacedName{Name: externalService.Name, Namespace: externalService.Namespace}

	if _, ok := p.probes[key]; ok {
		p.logger.Info("Removing probes", "externalservice", externalService.Name)
		p.RemoveProbes(externalService)
	}

	p.logger.Info("Adding probes", "externalservice", externalService.Name)
	p.AddProbes(externalService)
}
