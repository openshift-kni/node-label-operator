package controllers

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift-kni/node-label-operator/api/v1beta1"
	. "github.com/openshift-kni/node-label-operator/pkg/test"
)

var _ = Describe("OwnedLabels controller", func() {

	var nodeMatching *v1.Node
	var labels *v1beta1.Labels
	var labelsDeletedByTest bool

	BeforeEach(func() {
		labelsDeletedByTest = false

		By("Creating nodes")
		nodeMatching = GetNode(NodeNameMatching)
		Expect(k8sClient.Create(context.Background(), nodeMatching)).Should(Succeed(), "nodeMatching should have been created")

		By("Creating a Labels CR")
		labels = GetLabels()
		Expect(k8sClient.Create(context.Background(), labels)).Should(Succeed(), "labels should have been created")
	})

	AfterEach(func() {
		By("Cleaning up nodes and labels")
		Expect(k8sClient.Delete(context.Background(), nodeMatching)).Should(Succeed(), "nodeMatching should have been deleted")
		if !labelsDeletedByTest {
			Expect(k8sClient.Delete(context.Background(), labels)).Should(Succeed(), "labels should have been deleted")
			By("Ensure Labels is deleted")
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(labels), labels)
				return err != nil && errors.IsNotFound(err)
			}, Timeout, Interval).Should(BeTrue(), "labels should be away")
		}
	})

	When("Creating OwnedLabels", func() {

		var ownedLabels *v1beta1.OwnedLabels

		AfterEach(func() {
			Expect(k8sClient.Delete(context.Background(), ownedLabels)).Should(Succeed(), "ownedLabels should have been deleted")
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(ownedLabels), ownedLabels)
				return err != nil && errors.IsNotFound(err)
			}, Timeout, Interval).Should(BeTrue(), "ownedlabels should be away")
		})

		It("Should delete uncovered labels", func() {

			By("Verifying that label was set on matching node")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
				GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
				val, ok := nodeMatching.Labels[LabelDomainName]
				return ok && val == LabelValue
			}, Timeout, Interval).Should(BeTrue(), "label should have been set")

			By("Deleting Labels")
			Expect(k8sClient.Delete(context.Background(), labels)).Should(Succeed())
			labelsDeletedByTest = true

			By("Verifying that label isn't deleted yet")
			Consistently(func() bool {
				Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
				GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
				val, ok := nodeMatching.Labels[LabelDomainName]
				return ok && val == LabelValue
			}, Timeout, Interval).Should(BeTrue(), "label should not be deleted yet")

			By("Creating OwnedLabels")
			ownedLabels = GetOwnedLabels()
			Expect(k8sClient.Create(context.Background(), ownedLabels)).Should(Succeed(), "ownedLabels should have been created")

			By("Verifying that label is deleted now")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
				GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
				_, ok := nodeMatching.Labels[LabelDomainName]
				return ok
			}, Timeout, Interval).Should(BeFalse(), "label should be deleted now")

		})

	})

})
