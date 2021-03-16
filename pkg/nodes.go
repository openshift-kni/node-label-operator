package pkg

import (
	"fmt"
	"regexp"

	"github.com/go-logr/logr"

	v1 "k8s.io/api/core/v1"

	"github.com/openshift-kni/node-label-operator/api/v1beta1"
)

// RemoveOwnedLabels removes all uncovered owned labels from the node and return true if the node was modified
func RemoveOwnedLabels(node *v1.Node, allOwnedLabels []v1beta1.OwnedLabels, allLabels []v1beta1.Labels, log logr.Logger) bool {
	// check if we have owned labels on the node
	log.Info("Checking owned labels", "node", node.Name)
	nodeModified := false
	for labelDomainName := range node.Labels {
		// check if we own this label
		for _, ownedLabel := range allOwnedLabels {
			if !IsOwnedLabel(labelDomainName, ownedLabel, log) {
				continue
			}
			// we own this label
			// check if it is still covered by a label rule
			if !IsCoveredByAll(node.Name, labelDomainName, allLabels, log) {
				// we need to remove the label
				log.Info("Deleting uncovered owned label")
				delete(node.Labels, labelDomainName)
				nodeModified = true
			}
		}
	}
	return nodeModified
}

// AddAllLabels adds the labels configured in the rules of the given Labels to the given node
func AddAllLabels(node *v1.Node, allLabels []v1beta1.Labels, log logr.Logger) bool {
	nodeModified := false
	for _, labels := range allLabels {
		nodeModified = AddLabels(node, labels, log) || nodeModified
	}
	return nodeModified
}

// AddLabels adds the labels configured in the rules of the given Labels to the given node
func AddLabels(node *v1.Node, labels v1beta1.Labels, log logr.Logger) bool {
	if !labels.GetDeletionTimestamp().IsZero() {
		return false
	}
	log.Info("Checking if labels need to be added to node", "node", node.Name, "label config", fmt.Sprintf("%+v", labels.Spec))
	nodeModified := false
	for _, nodeNamePattern := range labels.Spec.NodeNamePatterns {
		pattern := fmt.Sprintf("%s%s%s", "^", nodeNamePattern, "$")
		match, err := regexp.MatchString(pattern, node.Name)
		if err != nil {
			log.Error(err, "Invalid regular expression, moving on to next rule")
			continue
		}
		if !match {
			continue
		}
		// we have a match, add labels!
		for name, value := range labels.Spec.Labels {
			if val, ok := node.Labels[name]; !ok || val != value {
				log.Info("Adding label to node based on pattern", "labelName", name, "labelValue", value, "pattern", nodeNamePattern)
				if node.Labels == nil {
					node.Labels = map[string]string{}
				}
				node.Labels[name] = value
				nodeModified = true
			}
		}
	}
	return nodeModified
}
