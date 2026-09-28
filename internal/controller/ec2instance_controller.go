/*
Copyright 2026.

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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	computev1 "github.com/qoofa/ec2-operator/api/v1"
)

// Ec2instanceReconciler reconciles a Ec2instance object
type Ec2instanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=compute.minhaj.blog,resources=ec2instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compute.minhaj.blog,resources=ec2instances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=compute.minhaj.blog,resources=ec2instances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ec2instance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.25.0/pkg/reconcile
func (r *Ec2instanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := logf.FromContext(ctx)

	l.Info("=== RECONCILE LOOP STARTED ===", "namespace", req.Namespace, "name", req.Name)

	ec2Instance := &computev1.Ec2instance{}

	if err := r.Get(ctx, req.NamespacedName, ec2Instance); err != nil {
		if errors.IsNotFound(err) {
			l.Info("Instance Deleted. No need to reconcile")

			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if ec2Instance.Status.InstanceID != "" {
		l.Info("=== Requested Instace exists already so skipping ", "instanceId", ec2Instance.Status.InstanceID, " ===")
		return ctrl.Result{}, nil
	}

	l.Info("Creating new instance")

	l.Info("=== Adding Finalizer ===")
	ec2Instance.Finalizers = append(ec2Instance.Finalizers, "ec2instance.compute.minhaj.blog")
	if err := r.Update(ctx, ec2Instance); err != nil {
		l.Error(err, "Failed to add finalizer")

		return ctrl.Result{
			Requeue: true,
		}, err
	}
	l.Info("=== Added Finalizer ===")

	l.Info("=== Creating new Instance in Reconciler ===")

	newInstance, err := createEc2Instance(ec2Instance)
	if err != nil {
		l.Error(err, "Failed to create EC2 instance")
		return ctrl.Result{}, nil
	}

	l.Info("=== updating status , would cause reconciler to be ran again ===",
		"instanceId", newInstance.InstanceID,
		"state", newInstance.State)

	ec2Instance.Status.InstanceID = newInstance.InstanceID
	ec2Instance.Status.State = newInstance.State
	ec2Instance.Status.PublicDNS = newInstance.PublicIP
	ec2Instance.Status.PrivateIP = newInstance.PrivateIP
	ec2Instance.Status.PublicDNS = newInstance.PrivateDNS
	ec2Instance.Status.PrivateDNS = newInstance.PrivateDNS

	err = r.Status().Update(ctx, ec2Instance)
	if err != nil {
		l.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}
	l.Info("=== Status Updated - Reconsiler would be retriggered ===")

	return ctrl.Result{
		RequeueAfter: time.Second * 1,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Ec2instanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1.Ec2instance{}).
		Named("ec2instance").
		Complete(r)
}
