/*
Copyright 2023.

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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	webappv1 "my.domain/myservice/api/v1"
)

// MyServiceReconciler reconciles a MyService object
type MyServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.my.domain,resources=myservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.my.domain,resources=myservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.my.domain,resources=myservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *MyServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("reconciliation start")

	// Fetch the MyService instance
	var myService webappv1.MyService
	if err := r.Get(ctx, req.NamespacedName, &myService); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to fetch MyService")
			return ctrl.Result{}, err
		}
		// MyService not found, return without error
		return ctrl.Result{}, nil
	}

	// Create the Deployment object
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      myService.Name,
			Namespace: myService.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &myService.Spec.DeploymentReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": myService.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": myService.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "myservice",
							Image: myService.Spec.DeploymentImage,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							// Command: []string{"<command>"},
							// Args:    []string{"<args>"},
						},
					},
				},
			},
		},
	}

	// Set MyService instance as the owner and controller
	if err := ctrl.SetControllerReference(&myService, deployment, r.Scheme); err != nil {
		logger.Error(err, "unable to set controller reference for Deployment")
		return ctrl.Result{}, err
	}

	// Create the Deployment object if it doesn't already exist
	if err := r.Create(ctx, deployment); err != nil {
		if !errors.IsAlreadyExists(err) {
			logger.Error(err, "unable to create Deployment")
			return ctrl.Result{}, err
		}
	}

	// Create the Service object
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      myService.Name,
			Namespace: myService.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: myService.Spec.ServiceType,
			Selector: map[string]string{
				"app": myService.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					TargetPort: intstr.FromInt(80),
					Port:       80,
				},
			},
		},
	}

	// Set MyService instance as the owner and controller
	if err := ctrl.SetControllerReference(&myService, service, r.Scheme); err != nil {
		logger.Error(err, "unable to set controller reference for Service")
		return ctrl.Result{}, err
	}

	// Create the Service object if it doesn't already exist
	if err := r.Create(ctx, service); err != nil {
		if !errors.IsAlreadyExists(err) {
			logger.Error(err, "unable to create Service")
			return ctrl.Result{}, err
		}
	}

	logger.Info("reconciliation complete")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.MyService{}).
		Complete(r)
}
