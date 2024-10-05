package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	resourcesv1alpha1 "github.com/opsground/kchron/api/v1alpha1"
)

var setupLog = ctrl.Log.WithName("Action")

// CronRestartReconciler reconciles a CronRestart object
type CronRestartReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	CronJobs  map[string]*cron.Cron // Map to store cron jobs
	CronMutex sync.Mutex            // Mutex to ensure thread safety
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=resources.kchron.io,resources=cronrestarts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=resources.kchron.io,resources=cronrestarts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=resources.kchron.io,resources=cronrestarts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CronRestart object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
// Initialize the CronJobs map
func NewCronRestartReconciler(client client.Client, scheme *runtime.Scheme) *CronRestartReconciler {
	return &CronRestartReconciler{
		Client:   client,
		Scheme:   scheme,
		CronJobs: make(map[string]*cron.Cron),
	}
}

func (r *CronRestartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var cronRestart resourcesv1alpha1.CronRestart
	if err := r.Get(ctx, req.NamespacedName, &cronRestart); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// The resource was not found, which means it was deleted
			return r.handleDeletion(req.NamespacedName)
		}
		return ctrl.Result{}, err
	}

	// Generate a unique cronJobID for the resource
	cronJobID := fmt.Sprintf("cron-%s-%s", cronRestart.Namespace, cronRestart.Name)

	// Ensure thread safety for the map access
	r.CronMutex.Lock()
	defer r.CronMutex.Unlock()

	// If a cron job already exists for this resource, stop it before creating a new one
	if existingCron, exists := r.CronJobs[cronJobID]; exists {
		existingCron.Stop() // Stop the existing cron job
		delete(r.CronJobs, cronJobID)
	}

	// Create a new cron job for the current CronRestart resource
	newCronJob := cron.New()

	_, err := newCronJob.AddFunc(cronRestart.Spec.CronSchedule, func() {
		err := r.restartResources(ctx, cronRestart.Spec.Namespace, cronRestart.Spec.ResourceType, cronRestart.Spec.Resources)
		if err != nil {
			setupLog.Error(err, "Failed to restart resources", "CronRestart", cronRestart.Name)
		}
	})
	if err != nil {
		setupLog.Error(err, "Failed to schedule cron job", "CronRestart", cronRestart.Name)
		return ctrl.Result{}, err
	}

	// Start the cron job and store it in the map
	newCronJob.Start()
	r.CronJobs[cronJobID] = newCronJob

	setupLog.Info("Cron job scheduled", "CronRestart", cronRestart.Name, "CronJobID", cronJobID)

	return ctrl.Result{}, nil
}

func (r *CronRestartReconciler) handleDeletion(namespacedName types.NamespacedName) (ctrl.Result, error) {
	cronJobID := fmt.Sprintf("cron-%s-%s", namespacedName.Namespace, namespacedName.Name)

	r.CronMutex.Lock()
	defer r.CronMutex.Unlock()

	if existingCron, exists := r.CronJobs[cronJobID]; exists {
		existingCron.Stop() // Stop the existing cron job
		delete(r.CronJobs, cronJobID)
		setupLog.Info("Cron job removed due to resource deletion", "CronJobID", cronJobID)
	}

	return ctrl.Result{}, nil
}

func (r *CronRestartReconciler) restartResources(ctx context.Context, namespace, resourceType string, resources []string) error {
	var errs []error
	for _, resourceName := range resources {
		switch resourceType {
		case "Deployment":
			if err := r.restartDeployment(ctx, namespace, resourceName); err != nil {
				errs = append(errs, err)
			}
		case "StatefulSet":
			if err := r.restartStatefulSet(ctx, namespace, resourceName); err != nil {
				errs = append(errs, err)
			}
		case "DaemonSet":
			if err := r.restartDaemonSet(ctx, namespace, resourceName); err != nil {
				errs = append(errs, err)
			}
		default:
			errs = append(errs, fmt.Errorf("unsupported resource type: %s", resourceType))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to restart resources: %v", errs)
	}
	return nil
}

func (r *CronRestartReconciler) restartDeployment(ctx context.Context, namespace, resourceName string) error {
	var deployment appsv1.Deployment
	if err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: resourceName}, &deployment); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		setupLog.Info("Deployment not found", "Namespace", namespace, "Name", resourceName)
		return nil
	}

	// Initialize the annotations map if it's nil
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}

	// Add restart annotation
	restartTime := time.Now().Format(time.RFC3339)
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartTime

	// Create a patch to update the deployment
	patch := client.MergeFrom(deployment.DeepCopy())
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartTime

	// Attempt to patch the Deployment
	if err := r.Patch(ctx, &deployment, patch); err != nil {
		setupLog.Error(err, "Failed to patch Deployment", "Namespace", namespace, "Name", resourceName)
		return err
	}

	// Log the successful update
	setupLog.Info("Successfully restarted Deployment", "Namespace", namespace, "Name", resourceName)

	return nil
}

func (r *CronRestartReconciler) restartStatefulSet(ctx context.Context, namespace, resourceName string) error {
	// Fetch the StatefulSet
	var statefulSet appsv1.StatefulSet
	if err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: resourceName}, &statefulSet); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		setupLog.Info("StatefulSet not found", "Namespace", namespace, "Name", resourceName)
		return nil
	}

	// Restart the StatefulSet by updating the annotation
	if statefulSet.Spec.Template.Annotations == nil {
		statefulSet.Spec.Template.Annotations = make(map[string]string)
	}
	// Add restart annotation
	restartTime := time.Now().Format(time.RFC3339)
	statefulSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartTime

	// Create a patch to update the statefulSet
	patch := client.MergeFrom(statefulSet.DeepCopy())
	statefulSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartTime

	// Attempt to patch the statefulSet
	if err := r.Patch(ctx, &statefulSet, patch); err != nil {
		setupLog.Error(err, "Failed to patch StatefulSet", "Namespace", namespace, "Name", resourceName)
		return err
	}

	// Log the successful update
	setupLog.Info("Successfully restarted StatefulSet", "Namespace", namespace, "Name", resourceName)

	return nil
}

func (r *CronRestartReconciler) restartDaemonSet(ctx context.Context, namespace, resourceName string) error {
	// Fetch the DaemonSet
	var daemonSet appsv1.DaemonSet
	if err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: resourceName}, &daemonSet); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		setupLog.Info("DaemonSet not found", "Namespace", namespace, "Name", resourceName)
		return nil
	}

	// Restart the DaemonSet by updating the annotation
	if daemonSet.Spec.Template.Annotations == nil {
		daemonSet.Spec.Template.Annotations = make(map[string]string)
	}
	// Add restart annotation
	restartTime := time.Now().Format(time.RFC3339)
	daemonSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartTime

	// Create a patch to update the daemonSet
	patch := client.MergeFrom(daemonSet.DeepCopy())
	daemonSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartTime

	// Attempt to patch the daemonSet
	if err := r.Patch(ctx, &daemonSet, patch); err != nil {
		setupLog.Error(err, "Failed to patch DaemonSet", "Namespace", namespace, "Name", resourceName)
		return err
	}

	// Log the successful update
	setupLog.Info("Successfully restarted DaemonSet", "Namespace", namespace, "Name", resourceName)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronRestartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	setupLog.Info("Setting up controller")
	setupLog.Info("Controller setup completed")

	if r.CronJobs == nil {
		r.CronJobs = make(map[string]*cron.Cron)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcesv1alpha1.CronRestart{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				setupLog.Info("ObjectCreated", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace(), "object", e.Object)
				return true
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				setupLog.Info("ObjectUpdated", "name", e.ObjectNew.GetName(), "namespace", e.ObjectNew.GetNamespace(), "object", e.ObjectNew)
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				setupLog.Info("ObjectDeleted", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
				return true
			},
		}).
		Complete(r)
}
