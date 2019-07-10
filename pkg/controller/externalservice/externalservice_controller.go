package externalservice

import (
	"context"

	esov1alpha1 "github.com/CrowdfoxGmbH/external-service-operator/pkg/apis/eso/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/api/extensions/v1beta1"

	"github.com/CrowdfoxGmbH/external-service-operator/pkg/prober"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_externalservice")

// Add creates a new ExternalService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	client := mgr.GetClient()

	return &ReconcileExternalService{
		client:       client,
		scheme:       mgr.GetScheme(),
		probeManager: prober.NewProber(client),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("externalservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ExternalService
	err = c.Watch(&source.Kind{Type: &esov1alpha1.ExternalService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Endpoints{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &esov1alpha1.ExternalService{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &esov1alpha1.ExternalService{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &extv1.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &esov1alpha1.ExternalService{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileExternalService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileExternalService{}

// ReconcileExternalService reconciles a ExternalService object
type ReconcileExternalService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client       client.Client
	scheme       *runtime.Scheme
	probeManager *prober.ProbeManager
}

// Reconcile reads that state of the cluster for a ExternalService object and makes changes based on the state read
// and what is in the ExternalService.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileExternalService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ExternalService")

	// Fetch the ExternalService instance
	instance := &esov1alpha1.ExternalService{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("ExternalService got removed")
			r.probeManager.RemoveProbesByNamespacedName(request.NamespacedName)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if result, err := r.reconcileEndpoints(instance, reqLogger); err != nil {
		return result, err
	}
	if result, err := r.reconcileService(instance, reqLogger); err != nil {
		return result, err
	}
	if result, err := r.reconcileIngress(instance, reqLogger); err != nil {
		return result, err
	}

	r.probeManager.UpdateProbes(instance)

	return reconcile.Result{}, nil
}
