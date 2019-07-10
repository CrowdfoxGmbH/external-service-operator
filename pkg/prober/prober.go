package prober

import (
	esov1alpha1 "github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/probe/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

type externalServiceProber struct {
	workerLock      sync.RWMutex
	externalService *esov1alpha1.ExternalService
	workers         map[string]*worker
	httpprober      http.Prober
}

func (e *externalServiceProber) removeWorker(w *worker) {
	e.workerLock.Lock()
	defer e.workerLock.Unlock()

	delete(e.workers, w.ip)
}

func (e *externalServiceProber) shutdownAllWorkers() {
	for _, ip := range e.externalService.Spec.Ips {
		if worker, ok := e.workers[ip]; ok {
			worker.stop()
			// we don't have to delete the workers as they
			// are calling themself removeWorker() during stop procedure
		}
	}
}

func (e *externalServiceProber) addWorkers(client client.Client, probe corev1.Probe) {
	for _, ip := range e.externalService.Spec.Ips {
		worker := &worker{
			stopCh:         make(chan struct{}, 1),
			parent:         e,
			client:         client,
			namespacedName: types.NamespacedName{Name: e.externalService.Name, Namespace: e.externalService.Namespace},
			probe:          probe,
			ip:             ip,
		}
		go worker.run()
		e.workers[ip] = worker
	}
}
