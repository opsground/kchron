package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resourcesv1alpha1 "github.com/opsground/kchron/api/v1alpha1"
)

var setupLog = ctrl.Log.WithName("Action")

// CronRestartReconciler reconciles a CronRestart object
type CronRestartReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
func (r *CronRestartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var cronRestart resourcesv1alpha1.CronRestart
	if err := r.Get(ctx, req.NamespacedName, &cronRestart); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Parse and schedule the cron
	cronJob := cron.New()
	cronJob.AddFunc(cronRestart.Spec.CronSchedule, func() {
		// Call the restart logic
		r.restartResources(ctx, cronRestart.Spec.Namespace, cronRestart.Spec.ResourceType, cronRestart.Spec.Resources)
	})
	cronJob.Start()

	return ctrl.Result{}, nil
}

func (r *CronRestartReconciler) restartResources(ctx context.Context, namespace, resourceType string, resources []string) error {
	for _, resourceName := range resources {
		switch resourceType {
		case "Deployment":
			if err := r.restartDeployment(ctx, namespace, resourceName); err != nil {
				return err
			}
		case "StatefulSet":
			if err := r.restartStatefulSet(ctx, namespace, resourceName); err != nil {
				return err
			}
		case "DaemonSet":
			if err := r.restartDaemonSet(ctx, namespace, resourceName); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported resource type: %s", resourceType)
		}
	}
	return nil
}

func (r *CronRestartReconciler) restartDeployment(ctx context.Context, namespace, resourceName string) error {
	// Fetch the Deployment
	var deployment appsv1.Deployment
	if err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: resourceName}, &deployment); err != nil {
		setupLog.Error(err, "GetFailed", "Deployment", resourceName, "Namespace", namespace)
		return err
	}

	// Restart the Deployment by updating the annotation
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	if err := r.Update(ctx, &deployment); err != nil {
		setupLog.Error(err, "UpdateFailed", "Deployment", resourceName, "Namespace", namespace)
		return err
	}

	setupLog.Info("Restarted", "Deployment", resourceName, "Namespace", namespace)
	return nil
}

func (r *CronRestartReconciler) restartStatefulSet(ctx context.Context, namespace, resourceName string) error {
	// Fetch the StatefulSet
	var statefulSet appsv1.StatefulSet
	if err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: resourceName}, &statefulSet); err != nil {
		setupLog.Error(err, "GetFailed", "StatefulSet", resourceName, "Namespace", namespace)
		return err
	}

	// Restart the StatefulSet by updating the annotation
	if statefulSet.Spec.Template.Annotations == nil {
		statefulSet.Spec.Template.Annotations = make(map[string]string)
	}
	statefulSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	if err := r.Update(ctx, &statefulSet); err != nil {
		setupLog.Error(err, "UpdateFailed", "StatefulSet", resourceName, "Namespace", namespace)
		return err
	}

	setupLog.Info("Restarted", "StatefulSet", resourceName, "Namespace", namespace)
	return nil
}

func (r *CronRestartReconciler) restartDaemonSet(ctx context.Context, namespace, resourceName string) error {
	// Fetch the DaemonSet
	var daemonSet appsv1.DaemonSet
	if err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: resourceName}, &daemonSet); err != nil {
		setupLog.Error(err, "GetFailed", "DaemonSet", resourceName, "Namespace", namespace)
		return err
	}

	// Restart the DaemonSet by updating the annotation
	if daemonSet.Spec.Template.Annotations == nil {
		daemonSet.Spec.Template.Annotations = make(map[string]string)
	}
	daemonSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	if err := r.Update(ctx, &daemonSet); err != nil {
		setupLog.Error(err, "UpdateFailed", "DaemonSet", resourceName, "Namespace", namespace)
		return err
	}

	setupLog.Info("Restarted", "DaemonSet", resourceName, "Namespace", namespace)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronRestartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcesv1alpha1.CronRestart{}).
		Complete(r)
}
