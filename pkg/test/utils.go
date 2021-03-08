package test

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/openshift-kni/node-label-operator/api/v1beta1"
)

const (
	Timeout  = time.Second * 1
	Interval = time.Millisecond * 100

	NodeNamePattern  = "node-match-.*"
	NodeNameMatching = "node-match-yes"
	NodeNameNoMatch  = "node-no-match"

	LabelDomain        = "test.opneshift.io"
	LabelName          = "foo1"
	LabelValue         = "bar1"
	LabelNameNew       = "foo2"
	LabelValueNew      = "bar2"
	LabelDomainName    = LabelDomain + "/" + LabelName
	LabelDomainNameNew = LabelDomain + "/" + LabelNameNew
	Label              = LabelDomainName + "=" + LabelValue
	LabelNewValue      = LabelDomainName + "=" + LabelValueNew
	LabelNewName       = LabelDomainNameNew + "=" + LabelValue
)

var ctx = context.Background()

func GetNode(name string) *v1.Node {
	return &v1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func GetLabels() *v1beta1.Labels {
	return &v1beta1.Labels{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Labels",
			APIVersion: "node-labels.openshift.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-labels-",
			Namespace:    "default",
		},
		Spec: v1beta1.LabelsSpec{
			Rules: []v1beta1.Rule{
				{
					NodeNamePatterns: []string{NodeNamePattern},
					Labels:           []string{Label},
				},
			},
		},
	}
}

func GetOwnedLabels() *v1beta1.OwnedLabels {
	return &v1beta1.OwnedLabels{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OwnedLabels",
			APIVersion: "node-labels.openshift.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-ownedlabels-",
			Namespace:    "default",
		},
		Spec: v1beta1.OwnedLabelsSpec{
			Domain: pointer.StringPtr(LabelDomain),
		},
	}
}
