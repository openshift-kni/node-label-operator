package tests

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift-kni/node-label-operator/api/v1beta1"
	. "github.com/openshift-kni/node-label-operator/pkg/test"
)

// Note: this file hasn't the _test.go postfix because it is reused by e2e tests,
// and _test.go files are only compiled if their own package is under test.

var _ = Describe("OwnedLabels controller", func() {

	var nodeMatching *v1.Node
	var labels *v1beta1.Labels
	var labelsDeletedByTest bool
	var k8sClient client.Client

	BeforeEach(func() {

		k8sClient = *K8sClient // from test package
		labelsDeletedByTest = false

		nodes := FindWorkerNodes()
		nodeMatching = nodes[0]

		By("Creating a Labels CR")
		nodeNamePattern := GetPattern(nodeMatching.Name, "")
		labels = GetLabels(nodeNamePattern)
		Expect(k8sClient.Create(context.Background(), labels)).Should(Succeed(), "labels should have been created")
	})

	AfterEach(func() {
		By("Cleaning up nodes and labels")
		CleanupDummyNodes()

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
