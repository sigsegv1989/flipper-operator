/*
Copyright 2024 Abhijeet Rokade.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	flipperv1alpha1 "github.com/sigsegv1989/flipper-operator/api/v1alpha1"
)

// RollingUpdateReconciler reconciles a RollingUpdate object
type RollingUpdateReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=flipper.example.com,resources=rollingupdates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=flipper.example.com,resources=rollingupdates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=flipper.example.com,resources=rollingupdates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RollingUpdate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *RollingUpdateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("rollingupdate", req.NamespacedName)

	// Fetch the RollingUpdate CR instance
	rollingUpdate := &flipperv1alpha1.RollingUpdate{}
	err := r.Get(ctx, req.NamespacedName, rollingUpdate)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("RollingUpdate resource not found. Ignoring reconcile...")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch RollingUpdate")
		return ctrl.Result{}, err
	}
	log.V(1).Info("Successfully retrieved RollingUpdate resource", "rollingUpdate", rollingUpdate)

	interval, err := time.ParseDuration(rollingUpdate.Spec.Interval)
	if err != nil {
		log.Error(err, "Failed to parse interval duration", "interval", rollingUpdate.Spec.Interval)
		return ctrl.Result{}, err
	}
	log.V(1).Info("Successfully retrieved RollingUpdate interval", "interval", interval)

	now := time.Now()
	if rollingUpdate.Status.LastRolloutTime.Time.IsZero() ||
		now.Sub(rollingUpdate.Status.LastRolloutTime.Time) > interval {
		log.V(1).Info("Time to rolling restart resources", "lastRolloutTime", rollingUpdate.Status.LastRolloutTime, "now", now, "interval", interval)

		deployments, err := r.restartDeployments(ctx, req, rollingUpdate.Spec.MatchLabels)
		if err != nil {
			log.Error(err, "Failed to restart deployments")
			return ctrl.Result{}, err
		}

		rollingUpdate.Status.LastRolloutTime = metav1.Now()
		rollingUpdate.Status.Deployments = deployments
		err = r.Status().Update(ctx, rollingUpdate)
		if err != nil {
			log.Error(err, "Failed to update rollingUpdate status")
			return ctrl.Result{}, err
		}

		log.Info("Successfully rolling restarted resource and updated RollingUpdate status", "lastRolloutTime", rollingUpdate.Status.LastRolloutTime)
	}

	return ctrl.Result{RequeueAfter: interval}, nil
}

func (r *RollingUpdateReconciler) restartDeployments(ctx context.Context, req ctrl.Request, labels map[string]string) ([]string, error) {
	log := r.Log.WithValues("namespace", req.Namespace, "name", req.Name)

	log.V(1).Info("Listing deployments for rolling restart", "labels", labels)
	log.V(1).Info("Checking deployment labels", "namespace", req.Namespace, "expectedLabels", labels)

	deployments := &appsv1.DeploymentList{}
	err := r.List(ctx, deployments, client.InNamespace(req.Namespace), client.MatchingLabels(labels))
	if err != nil {
		log.Error(err, "Failed to list deployments", "labels", labels)
		return []string{}, err
	}
	log.V(1).Info("Deployments listed", "deploymentCount", len(deployments.Items))

	deploys := []string{}
	for _, deployment := range deployments.Items {
		deploys = append(deploys, deployment.Name)
		log.V(1).Info("Restarting deployment", "namespace", deployment.Namespace, "name", deployment.Name)

		annotations := map[string]string{
			"kubectl.kubernetes.io/restartedAt":      time.Now().Format(time.RFC3339),
			"kubectl.kubernetes.io/restartedBy":      "flipper-operator",
			"flipper.example.com/restartedByCR":      fmt.Sprintf("%s/%s", req.Namespace, req.Name),
			"flipper.example.com/restartedByCRDKind": "rollingupdate",
		}
		r.updateAnnotations(&deployment, annotations)

		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return r.Update(ctx, &deployment)
		})

		if err != nil {
			log.Error(err, "Failed to update deployment", "name", deployment.Name)
			return []string{}, fmt.Errorf("failed to update Deployment %s/%s: %v", deployment.Namespace, deployment.Name, err)
		}

		log.Info("Successfully rolling restarted deployment", "name", deployment.Name)
	}

	return deploys, nil
}

func (r *RollingUpdateReconciler) updateAnnotations(deployment *appsv1.Deployment, annotations map[string]string) {
	// Update the Deployment's annotations
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	for key, value := range annotations {
		deployment.Annotations[key] = value
	}

	// Update the pod template spec's annotations to trigger a rollout
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	for key, value := range annotations {
		deployment.Spec.Template.Annotations[key] = value
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RollingUpdateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log = mgr.GetLogger().WithName("controller").WithName("RollingUpdate")

	return ctrl.NewControllerManagedBy(mgr).
		For(&flipperv1alpha1.RollingUpdate{}).
		Complete(r)
}
