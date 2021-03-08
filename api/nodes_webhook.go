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
package api

import (
	"context"
	"encoding/json"
	"net/http"

	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/openshift-kni/node-label-operator/api/v1beta1"
	"github.com/openshift-kni/node-label-operator/pkg"
)

// +kubebuilder:webhook:path=/label-v1-nodes,mutating=true,failurePolicy=ignore,sideEffects=None,groups="",resources=nodes,verbs=create,versions=v1,name=mnode.kb.io,admissionReviewVersions={v1,v1beta1}

// log is for logging in this package.
var log = logf.Log.WithName("nodes-webhook")

// NodeLabeler adds labels to Nodes
type NodeLabeler struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (n *NodeLabeler) Handle(ctx context.Context, req admission.Request) admission.Response {

	log.Info("node webhook is called!")

	node := &v1.Node{}
	err := n.decoder.Decode(req, node)
	if err != nil {
		log.Error(err, "Failed to decode node")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// get all label rules and apply labels as they match
	allLabels := &v1beta1.LabelsList{}
	if err = n.Client.List(context.TODO(), allLabels, &client.ListOptions{}); err != nil {
		log.Error(err, "Failed to list Labels")
		return admission.Errored(http.StatusBadRequest, err)
	}

	nodeModified := pkg.AddAllLabels(node, allLabels.Items, log)

	if nodeModified {
		marshaledNode, err := json.Marshal(node)
		if err != nil {
			log.Error(err, "marshalling response went wrong")
			return admission.Errored(http.StatusInternalServerError, err)
		}
		return admission.PatchResponseFromRaw(req.Object.Raw, marshaledNode)
	}

	return admission.Allowed("no label added")
}

// InjectDecoder injects the decoder.
func (n *NodeLabeler) InjectDecoder(d *admission.Decoder) error {
	n.decoder = d
	return nil
}

func (n *NodeLabeler) SetupWebhookWithManager(mgr ctrl.Manager) {
	hookServer := mgr.GetWebhookServer()
	hookServer.Register("/label-v1-nodes", &webhook.Admission{Handler: &NodeLabeler{Client: mgr.GetClient()}})
}
