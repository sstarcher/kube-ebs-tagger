package persistentvolume

import (
	"context"
	"fmt"
	"strings"

	"github.com/sstarcher/kube-ebs-tagger/pkg/tagger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	// "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_persistentvolume")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new PersistentVolume Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePersistentVolume{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("persistentvolume-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PersistentVolume
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner PersistentVolume
	// fmt.Println("setting up watch")
	// err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
	// 	IsController: true,
	// })
	// if err != nil {
	// 	return err
	// }

	return nil
}

// blank assignment to verify that ReconcilePersistentVolume implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePersistentVolume{}

// ReconcilePersistentVolume reconciles a PersistentVolume object
type ReconcilePersistentVolume struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PersistentVolume object and makes changes based on the state read
// and what is in the PersistentVolume.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePersistentVolume) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PersistentVolume")

	// Fetch the PersistentVolume instance
	instance := &corev1.PersistentVolumeClaim{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	metadata := instance.GetObjectMeta()
	annotations := metadata.GetAnnotations()

	if annotations["volume.beta.kubernetes.io/storage-provisioner"] == "kubernetes.io/aws-ebs" {
		if instance.Spec.VolumeName != "" {

			pv := &corev1.PersistentVolume{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.VolumeName}, pv)
			if err != nil && errors.IsNotFound(err) {
				reqLogger.Info("PV does not exist yet", "PV.Name", instance.Spec.VolumeName)
				return reconcile.Result{}, nil
			}

			data := strings.Split(pv.Spec.AWSElasticBlockStore.VolumeID, "/")
			volumeID := data[0]
			if len(data) == 4 {
				volumeID = data[3]
			}
			if !strings.HasPrefix(volumeID, "vol-") {
				err := fmt.Errorf("volume not in normal format got %s", volumeID)
				reqLogger.Error(err, "expected to start with vol-")
			}

			labels := map[string]string{}
			for k, v := range metadata.GetLabels() {
				labels[k] = v
			}

			for k, v := range pv.GetObjectMeta().GetLabels() {
				labels[k] = v
			}

			updated, err := tagger.Tag(volumeID, labels)
			if err != nil {
				reqLogger.Error(err, "")
			}
			if updated {
				reqLogger.Info("updated tags")
			}

		} else {
			reqLogger.Info("not a aws volume")
		}
	} else {
		reqLogger.Info("not a aws-ebs provisioner")
	}

	return reconcile.Result{}, nil
}
