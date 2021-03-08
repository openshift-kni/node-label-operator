package pkg

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-logr/logr"

	"github.com/openshift-kni/node-label-operator/api/v1beta1"
)

// IsOwnedLabel checks if the given nodeLabelDomainName matches the given ownedLabel
func IsOwnedLabel(nodeLabelDomainName string, ownedLabel v1beta1.OwnedLabels, log logr.Logger) bool {
	log.Info("Check if we own label", "labelDomainName", nodeLabelDomainName, "OwnedLabel", ownedLabel.Name)

	// split domainName
	parts := strings.Split(nodeLabelDomainName, "/")
	if len(parts) != 2 {
		// this should not happen...
		log.Info("Skipping unexpected label", "labelDomainName", nodeLabelDomainName)
		return false
	}
	labelDomain := parts[0]
	labelName := parts[1]

	// check if we own this label
	if ownedLabel.Spec.Domain != nil && *ownedLabel.Spec.Domain != labelDomain {
		// domain set but doesn't match, move on
		log.Info("Domain does not match", "nodeLabelDomain", labelDomain, "ownedLabelDomain", ownedLabel.Spec.Domain)
		return false
	}
	if ownedLabel.Spec.NamePattern != nil {
		pattern := fmt.Sprintf("%s%s%s", "^", *ownedLabel.Spec.NamePattern, "$")
		match, err := regexp.MatchString(pattern, labelName)
		if err != nil {
			log.Error(err, "Invalid regular expression, moving on", "pattern", ownedLabel.Spec.NamePattern)
			return false
		}
		if !match {
			// name pattern set but doesn't match, move on
			log.Info("Name pattern does not match", "nodeLabelName", labelName, "ownedLabelNamePattern", ownedLabel.Spec.NamePattern)
			return false
		}
	}

	log.Info("We own it!")
	return true
}
