package prober

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/kubernetes/pkg/probe"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("worker")

type worker struct {
	stopCh          chan struct{}
	parent          *externalServiceProber
	namespacedName  types.NamespacedName
	client          client.Client
	probe           corev1.Probe
	ip              string
	lastResultType  probe.Result
	lastResultCount int32
}

func (w *worker) run() {
	probeTickerPeriod := time.Duration(w.probe.PeriodSeconds) * time.Second
	w.lastResultType = probe.Unknown

	// If kubelet restarted the probes could be started in rapid succession.
	// Let the worker wait for a random portion of tickerPeriod before probing.
	time.Sleep(time.Duration(rand.Float64()*float64(probeTickerPeriod) + float64(w.probe.InitialDelaySeconds)))

	probeTicker := time.NewTicker(probeTickerPeriod)

	defer func() {
		// Clean up.
		probeTicker.Stop()
		log.Info("Stop Prober", "endpoint", w.parent.externalService.Name, "IP", w.ip)
		w.parent.removeWorker(w)
	}()

	log.Info("Start Prober", "endpoint", w.parent.externalService.Name, "IP", w.ip)
probeLoop:
	for w.doProbe() {
		// Wait for next probe tick.
		select {
		case <-w.stopCh:
			log.V(1).Info("Received Stop", "endpoint", w.parent.externalService.Name, "IP", w.ip)
			break probeLoop
		case <-probeTicker.C:
			// continue
		}
	}
}

func (w *worker) stop() {
	log.V(1).Info("Sending Stop Signal", "endpoint", w.parent.externalService.Name, "ip", w.ip)
	select {
	case w.stopCh <- struct{}{}:
	default:
	}
}

func (w *worker) doProbe() (keepGoing bool) {

	runLogger := log.WithValues("IP", w.ip, "Port", w.probe.HTTPGet.Port, "endpoint", w.namespacedName.Name, "namespace", w.namespacedName.Namespace)

	runLogger.V(1).Info("Start Check")
	defer func() { recover() }() // Actually eat panics (HandleCrash takes care of logging)
	defer runtime.HandleCrash(func(_ interface{}) { keepGoing = true })

	endpoint := &corev1.Endpoints{}
	err := w.client.Get(context.TODO(), w.namespacedName, endpoint)

	if err != nil {
		if kerrors.IsNotFound(err) {
			// Endpoint could not be found, so we can stop monitor it
			runLogger.Error(err, "Endpoint could not be found")
			return false
		} else {
			//If its an unknown error, we should just skip healthchecking
			runLogger.Error(err, "Error getting Endpoint resource")
			return true
		}
	}

	result, message, err := w.runHttpProbe()

	if err != nil {
		runLogger.Error(err, "Runtimeerror during probe", "message", message)
		// TODO: add Note to ExternalServiceRessource that something is wrong with the Probe
		return false
	}

	if w.lastResultType == result {
		//prevent overflow
		if w.lastResultCount < math.MaxUint16 {
			w.lastResultCount++
		}
	} else {
		w.lastResultType = result
		w.lastResultCount = 1
	}

	runLogger.V(1).Info("Increased Last ResultCount", "type", w.lastResultType, "count", w.lastResultCount)

	switch result {
	case probe.Success:
		if w.lastResultCount >= w.probe.SuccessThreshold {
			if err := w.ensureReady(endpoint); err != nil {
				runLogger.Error(err, "Couldn't update Endpoint Ressource.", "endpoint", endpoint)
				return true //keep checking
			}

		}
	case probe.Failure:
		if w.lastResultCount >= w.probe.FailureThreshold {
			if err := w.ensureUnready(endpoint); err != nil {
				runLogger.Error(err, "Couldn't update Endpoint Ressource", "endpoint", endpoint)
				return true //keep checking
			}
		}
	case probe.Unknown:
		runLogger.Error(nil, "Health of Endpoint is unkown")
	}

	return true
}

func (w *worker) runHttpProbe() (probe.Result, string, error) {
	scheme := strings.ToLower(string(w.probe.HTTPGet.Scheme))
	port := w.probe.HTTPGet.Port.IntValue()
	path := w.probe.HTTPGet.Path
	url := formatURL(scheme, w.ip, port, path)
	headers := buildHeader(w.probe.HTTPGet)
	timeout := time.Duration(w.probe.TimeoutSeconds) * time.Second

	return w.parent.httpprober.Probe(url, headers, timeout)
}

func buildHeader(httpaction *corev1.HTTPGetAction) http.Header {
	headers := make(http.Header)

	if httpaction.Host != "" {
		headers.Set("Host", httpaction.Host)
	}

	for _, header := range httpaction.HTTPHeaders {
		headers[header.Name] = append(headers[header.Name], header.Value)
	}

	return headers
}

func formatURL(scheme string, host string, port int, path string) *url.URL {
	u, err := url.Parse(path)
	// Something is busted with the path, but it's too late to reject it. Pass it along as is.
	if err != nil {
		u = &url.URL{
			Path: path,
		}
	}
	u.Scheme = scheme
	u.Host = net.JoinHostPort(host, strconv.Itoa(port))
	return u
}

func (w *worker) ensureReady(endpoint *corev1.Endpoints) error {
	newNotReadyAddresses := []corev1.EndpointAddress{}
	newReadyAddresses := endpoint.Subsets[0].Addresses

	found := false

	for _, address := range endpoint.Subsets[0].NotReadyAddresses {
		if address.IP != w.ip {
			newNotReadyAddresses = append(newNotReadyAddresses, address)
		} else {
			newReadyAddresses = append(newReadyAddresses, address)
			found = true
		}
	}

	for _, address := range endpoint.Subsets[0].Addresses {
		if address.IP == w.ip {
			found = true
			break
		}
	}

	if !found {
		return errors.New("couldn't find endpoint while marking it to ready")
	}

	return w.updateEndpoint(endpoint, newNotReadyAddresses, newReadyAddresses)
}

func (w *worker) ensureUnready(endpoint *corev1.Endpoints) error {
	newNotReadyAddresses := endpoint.Subsets[0].NotReadyAddresses
	newReadyAddresses := []corev1.EndpointAddress{}

	found := false

	for _, address := range endpoint.Subsets[0].Addresses {
		if address.IP != w.ip {
			newReadyAddresses = append(newReadyAddresses, address)
		} else {
			newNotReadyAddresses = append(newNotReadyAddresses, address)
			found = true
		}
	}

	for _, address := range endpoint.Subsets[0].NotReadyAddresses {
		if address.IP == w.ip {
			found = true
			break
		}
	}

	if !found {
		return errors.New("couldn't find endpoint while marking it to unready")
	}

	return w.updateEndpoint(endpoint, newNotReadyAddresses, newReadyAddresses)
}

func (w *worker) updateEndpoint(endpoint *corev1.Endpoints, notReady []corev1.EndpointAddress, ready []corev1.EndpointAddress) error {
	changed := !addressesEqual(endpoint.Subsets[0].NotReadyAddresses, notReady) || !addressesEqual(endpoint.Subsets[0].Addresses, ready)

	if changed {
		endpoint.Subsets[0].NotReadyAddresses = notReady
		endpoint.Subsets[0].Addresses = ready
		log.Info("Update Endpoint, because availability changed", "endpoint", endpoint.Name, "namespace", endpoint.Namespace)
		return w.client.Update(context.TODO(), endpoint)
	}
	return nil
}

// This function asserts that both arrays are sorted
func addressesEqual(a []corev1.EndpointAddress, b []corev1.EndpointAddress) bool {
	if cap(a) != cap(b) {
		return false
	}

	for i, valueA := range a {
		valueB := b[i]

		if valueB != valueA {
			return false
		}
	}
	return true
}
