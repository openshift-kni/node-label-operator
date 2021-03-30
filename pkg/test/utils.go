package test

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift-kni/node-label-operator/api/v1beta1"
)

const (
	Timeout  = time.Second * 5
	Interval = time.Millisecond * 500

	DummyNode1Name = "first-node"
	DummyNode2Name = "second-node"

	LabelDomain        = "test.openshift.io"
	LabelName          = "foo1"
	LabelValue         = "bar1"
	LabelNameNew       = "foo2"
	LabelValueNew      = "bar2"
	LabelDomainName    = LabelDomain + "/" + LabelName
	LabelDomainNameNew = LabelDomain + "/" + LabelNameNew
)

var (
	Label         = map[string]string{LabelDomainName: LabelValue}
	LabelNewValue = map[string]string{LabelDomainName: LabelValueNew}
	LabelNewName  = map[string]string{LabelDomainNameNew: LabelValue}

	K8sClient *client.Client
	IsE2etest = false
)

func GetLabels(nodeNamePattern string) *v1beta1.Labels {
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
			NodeNamePatterns: []string{nodeNamePattern},
			Labels:           Label,
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

func FindWorkerNodes() []*v1.Node {

	var nodes []*v1.Node

	if IsE2etest {

		// in real clusters dummy nodes don't live long enough for doing several tests on them
		// so use existing nodes
		By("Getting cluster nodes")
		nodeList := &v1.NodeList{}
		ExpectWithOffset(1, (*K8sClient).List(context.Background(), nodeList)).To(Succeed())
		for i, node := range nodeList.Items {
			if _, ok := node.Labels["node-role.kubernetes.io/worker"]; ok {
				nodes = append(nodes, &nodeList.Items[i])
			}
		}
		ExpectWithOffset(1, len(nodes)).To(BeNumerically(">=", 2), "didn't find enough worker nodes")
		return nodes

	}

	By("Creating dummy nodes")
	n1 := GetNode(DummyNode1Name)
	Expect((*K8sClient).Create(context.Background(), n1)).Should(Succeed(), "1st dummy should have been created")
	n2 := GetNode(DummyNode2Name)
	Expect((*K8sClient).Create(context.Background(), n2)).Should(Succeed(), "2nd dummy should have been created")
	nodes = append(nodes, n1, n2)
	return nodes

}

func CleanupDummyNodes() {
	if !IsE2etest {
		By("Cleaning up nodes and labels")
		Expect((*K8sClient).Delete(context.Background(), GetNode(DummyNode1Name))).Should(Succeed(), "dummy node 1 should have been deleted")
		Expect((*K8sClient).Delete(context.Background(), GetNode(DummyNode2Name))).Should(Succeed(), "dummy node 2 should have been deleted")
	}
}

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

func GetPattern(match, noMatch string) string {
	// create a pattern by replacing some chars with a wildcard, and ensure that it doesn't match "noMatch" if given
	stringLength := int(math.Min(float64(len(match)), float64(len(noMatch))))
	wildcardLength := int(stringLength / 2)
	start := 2
	var pattern string
	for {
		pattern = fmt.Sprintf("%s.*%s", match[:start], match[start+wildcardLength:])
		if noMatch == "" {
			break
		}
		matched, err := regexp.MatchString(pattern, noMatch)
		if err == nil && !matched {
			break
		}
		start++
		ExpectWithOffset(1, start+wildcardLength).To(BeNumerically("<=", stringLength), "didn't find a pattern!")
	}
	By("Using pattern " + pattern)
	return pattern
}
