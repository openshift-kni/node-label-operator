/*
Copyright 2021.

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

	"github.com/go-logr/logr"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/openshift-kni/node-label-operator/api/v1beta1"
	"github.com/openshift-kni/node-label-operator/pkg"
)

const labelsFinalizer = "node-label-operator.openshift.io/finalizer"

// LabelsReconciler reconciles a Labels object
type LabelsReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=node-labels.openshift.io,resources=labels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=node-labels.openshift.io,resources=labels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=node-labels.openshift.io,resources=labels/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Labels object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *LabelsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("labels", req.NamespacedName)

	log.Info("Reconciling")

	// get Labels instance
	labels := &v1beta1.Labels{}
	err := r.Get(ctx, req.NamespacedName, labels)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			log.Info("Labels resource not found, ignoring because it must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Labels")
		return ctrl.Result{}, err
	}

	markedForDeletion := !labels.GetDeletionTimestamp().IsZero()

	// add finalizer
	if !markedForDeletion && !controllerutil.ContainsFinalizer(labels, labelsFinalizer) {
		log.Info("adding finalizer")
		controllerutil.AddFinalizer(labels, labelsFinalizer)
		err = r.Update(ctx, labels)
		if err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// iterate all nodes
	// we have to
	// - remove all owned labels, if they aren't in any label rule
	// - add labels of this instance

	// we need all Labels
	allLabels := &v1beta1.LabelsList{}
	if err = r.Client.List(context.TODO(), allLabels, &client.ListOptions{}); err != nil {
		log.Error(err, "Failed to list Labels")
		return ctrl.Result{}, err
	}

	// and OwnedLabels
	ownedLabels := &v1beta1.OwnedLabelsList{}
	if err = r.Client.List(context.TODO(), ownedLabels, &client.ListOptions{}); err != nil {
		log.Error(err, "Failed to list OwnedLabels")
		return ctrl.Result{}, err
	}

	// get nodes
	nodes := &v1.NodeList{}
	if err = r.Client.List(context.TODO(), nodes, &client.ListOptions{}); err != nil {
		log.Error(err, "Failed to list Nodes")
		return ctrl.Result{}, err
	}

	// and start
	for _, nodeOrig := range nodes.Items {

		log.Info("Processing node", "nodeName", nodeOrig.Name)

		node := nodeOrig.DeepCopy()
		nodeModified := pkg.RemoveOwnedLabels(node, ownedLabels.Items, allLabels.Items, log)

		// owned labels are removed now on this node
		// add new / modified labels
		nodeModified = pkg.AddLabels(node, *labels, log) || nodeModified

		// save node
		if nodeModified {
			log.Info("patching node")
			baseToPatch := client.MergeFrom(&nodeOrig)
			if err := r.Client.Patch(context.TODO(), node, baseToPatch); err != nil {
				log.Error(err, "Failed to patch Node")
				return ctrl.Result{}, err
			}
		}

	}

	// remove finalizer
	if markedForDeletion {
		log.Info("removing finalizer")
		controllerutil.RemoveFinalizer(labels, labelsFinalizer)
		err := r.Update(ctx, labels)
		if err != nil {
			log.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LabelsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Labels{}).
		Complete(r)
}
