package pkg

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-logr/logr"

	"github.com/openshift-kni/node-label-operator/api/v1beta1"
)

// IsCoveredByAll checks if the given labelDomainName is covered by the rules of the given allLabels for the given nodeName
func IsCoveredByAll(nodeName string, labelDomainName string, allLabels []v1beta1.Labels, log logr.Logger) bool {
	for _, labels := range allLabels {
		if IsCovered(nodeName, labelDomainName, labels, log) {
			return true
		}
	}
	return false
}

// IsCoveredByAll checks if the given labelDomainName is covered by the rules of the given labels for the given nodeName
func IsCovered(nodeName string, labelDomainName string, labels v1beta1.Labels, log logr.Logger) bool {

	if !labels.GetDeletionTimestamp().IsZero() {
		return false
	}

	log.Info("Checking if label is covered", "node", nodeName, "label to check", labelDomainName, "label config", fmt.Sprintf("%+v", labels.Spec.Rules))
	for _, rule := range labels.Spec.Rules {
		for _, ruleLabel := range rule.Labels {
			// split to domain/name and value
			parts := strings.Split(ruleLabel, "=")
			if len(parts) != 2 {
				log.Info("Skipping unexpected rule label", ruleLabel)
				continue
			}
			if parts[0] == labelDomainName {
				log.Info("Label name matches")
				// label matches... does the node?
				for _, nodeNamePattern := range rule.NodeNamePatterns {
					match, err := regexp.MatchString(nodeNamePattern, nodeName)
					if err != nil {
						log.Error(err, "Invalid regular expression, moving on to next rule")
						continue
					}
					if match {
						// label is covered
						log.Info("Label value matches")
						return true
					}
				}
			}
		}
	}
	return false
}
