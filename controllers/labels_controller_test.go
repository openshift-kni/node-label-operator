package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/openshift-kni/node-label-operator/api/v1beta1"
	. "github.com/openshift-kni/node-label-operator/pkg/test"
)

var _ = Describe("Labels controller", func() {

	var nodeNotMatching *v1.Node
	var nodeMatching *v1.Node
	var labels *v1beta1.Labels
	var labelsDeletedByTest bool

	BeforeEach(func() {
		labelsDeletedByTest = false

		By("Creating nodes")
		nodeNotMatching = GetNode(NodeNameNoMatch)
		Expect(k8sClient.Create(context.Background(), nodeNotMatching)).Should(Succeed(), "nodeNotMatching should have been created")
		nodeMatching = GetNode(NodeNameMatching)
		Expect(k8sClient.Create(context.Background(), nodeMatching)).Should(Succeed(), "nodeMatching should have been created")

		By("Creating a Labels CR")
		labels = GetLabels()
		Expect(k8sClient.Create(context.Background(), labels)).Should(Succeed(), "labels should have been created")
	})

	AfterEach(func() {
		By("Cleaning up nodes and labels")
		Expect(k8sClient.Delete(context.Background(), nodeNotMatching)).Should(Succeed(), "nodeNotMatching should have been deleted")
		Expect(k8sClient.Delete(context.Background(), nodeMatching)).Should(Succeed(), "nodeMatching should have been deleted")
		if !labelsDeletedByTest {
			Expect(k8sClient.Delete(context.Background(), labels)).Should(Succeed(), "labels should have been deleted")
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(labels), labels)
				return err != nil && errors.IsNotFound(err)
			}, Timeout, Interval).Should(BeTrue(), "labels should be away")
		}
	})

	When("Creating a Labels CR", func() {
		It("Should add label to matching node", func() {

			By("Verifying that label was set on matching node")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
				GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
				val, ok := nodeMatching.Labels[LabelDomainName]
				return ok && val == LabelValue
			}, Timeout, Interval).Should(BeTrue(), "label should have been set")

			By("Verifying that label was not set on not matching node")
			Consistently(func() bool {
				Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeNotMatching), nodeNotMatching)).Should(Succeed())
				GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeNotMatching.Labels)))
				_, ok := nodeNotMatching.Labels[LabelDomainName]
				return ok
			}, Timeout, Interval).Should(BeFalse(), "label should not have been set")

		})
	})

	When("Updating a Labels CR", func() {

		Context("Without OwnedLabels", func() {

			It("Should update label value on matching node", func() {

				By("get latest labels version")
				labelsOrig := labels.DeepCopy()
				labels.Spec.Labels = LabelNewValue
				Expect(k8sClient.Patch(context.Background(), labels, client.MergeFrom(labelsOrig))).Should(Succeed())

				By("Verifying that label with updated value exists")
				Eventually(func() bool {
					Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
					GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
					val, ok := nodeMatching.Labels[LabelDomainName]
					return ok && val == LabelValueNew
				}, Timeout, Interval).Should(BeTrue(), "label should have been updated (new value exists)")

				By("Verifying that label with old value doesn't exists")
				Consistently(func() bool {
					Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
					GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
					val, ok := nodeMatching.Labels[LabelDomainName]
					return ok && val == LabelValue
				}, Timeout, Interval).Should(BeFalse(), "label should have been updated (old value doesn't exist)")

			})

			It("Should not update label name on matching node, but create a new", func() {

				By("Verifying that label was set on matching node")
				Eventually(func() bool {
					Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
					GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
					val, ok := nodeMatching.Labels[LabelDomainName]
					return ok && val == LabelValue
				}, Timeout, Interval).Should(BeTrue(), "label should have been set")

				By("Patching label name")
				labelsOrig := labels.DeepCopy()
				labels.Spec.Labels = LabelNewName
				Expect(k8sClient.Patch(context.Background(), labels, client.MergeFrom(labelsOrig))).Should(Succeed())

				By("Verifying that old label exists")
				Eventually(func() bool {
					Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
					GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
					val, ok := nodeMatching.Labels[LabelDomainName]
					return ok && val == LabelValue
				}, Timeout, Interval).Should(BeTrue(), "old label should still exist")

				By("Verifying that new label exists")
				Eventually(func() bool {
					Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
					GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
					val, ok := nodeMatching.Labels[LabelDomainNameNew]
					return ok && val == LabelValue
				}, Timeout, Interval).Should(BeTrue(), "new label should have been set")

			})

		})

		Context("With OwnedLabels", func() {

			var ownedLabels *v1beta1.OwnedLabels

			BeforeEach(func() {
				ownedLabels = GetOwnedLabels()
				Expect(k8sClient.Create(context.Background(), ownedLabels)).Should(Succeed(), "ownedLabels should have been created")
			})

			AfterEach(func() {
				Expect(k8sClient.Delete(context.Background(), ownedLabels)).Should(Succeed(), "ownedLabels should have been deleted")
				Eventually(func() bool {
					err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(ownedLabels), ownedLabels)
					return err != nil && errors.IsNotFound(err)
				}, Timeout, Interval).Should(BeTrue(), "ownedlabels should be away")
			})

			It("Should update label name on matching node", func() {

				By("Verifying that label was set on matching node")
				Eventually(func() bool {
					Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
					GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
					val, ok := nodeMatching.Labels[LabelDomainName]
					return ok && val == LabelValue
				}, Timeout, Interval).Should(BeTrue(), "label should have been set")

				By("Patching label name")
				labelsOrig := labels.DeepCopy()
				labels.Spec.Labels = LabelNewName
				Expect(k8sClient.Patch(context.Background(), labels, client.MergeFrom(labelsOrig))).Should(Succeed())

				By("Verifying that new label exists")
				Eventually(func() bool {
					Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
					GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
					val, ok := nodeMatching.Labels[LabelDomainNameNew]
					return ok && val == LabelValue
				}, Timeout, Interval).Should(BeTrue(), "new label should exist")

				By("Verifying that old label doesn't exists")
				Consistently(func() bool {
					Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
					GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
					_, ok := nodeMatching.Labels[LabelDomainName]
					return ok
				}, Timeout, Interval).Should(BeFalse(), "old label should be deleted")

			})

		})

	})

	When("Deleting a Labels CR", func() {

		Context("Without OwnedLabels", func() {

			It("Should not delete label on matching node", func() {

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
				By("Ensure Labels is deleted")
				Eventually(func() bool {
					err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(labels), labels)
					return err != nil && errors.IsNotFound(err)
				}, Timeout, Interval).Should(BeTrue(), "labels should be away")

				list := &v1beta1.OwnedLabelsList{}
				Expect(k8sClient.List(context.Background(), list)).To(Succeed())
				logf.Log.Info(fmt.Sprintf("OwnedLabels %+v", list))
				By("Verifying labels still exists")
				Consistently(func() bool {
					Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
					GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
					val, ok := nodeMatching.Labels[LabelDomainName]
					return ok && val == LabelValue
				}, Timeout, Interval).Should(BeTrue(), "label should not be deleted")
			})

		})

		Context("With OwnedLabels", func() {

			var ownedLabels *v1beta1.OwnedLabels

			BeforeEach(func() {
				ownedLabels = GetOwnedLabels()
				Expect(k8sClient.Create(context.Background(), ownedLabels)).Should(Succeed(), "ownedLabels should have been created")
			})

			AfterEach(func() {
				Expect(k8sClient.Delete(context.Background(), ownedLabels)).Should(Succeed(), "ownedLabels should have been deleted")
				Eventually(func() bool {
					err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(ownedLabels), ownedLabels)
					return err != nil && errors.IsNotFound(err)
				}, Timeout, Interval).Should(BeTrue(), "ownedlabels should be away")
			})

			It("Should delete label on matching node", func() {

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

				By("Verifying label was deleted")
				Eventually(func() bool {
					Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(nodeMatching), nodeMatching)).Should(Succeed())
					GinkgoWriter.Write([]byte(fmt.Sprintf("labels: %+v\n", nodeMatching.Labels)))
					val, ok := nodeMatching.Labels[LabelDomainName]
					return ok && val == LabelValue
				}, Timeout, Interval).Should(BeFalse(), "label should be deleted")

			})

		})

	})

})
