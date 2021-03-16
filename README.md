# Node Label Operator

## Purpose

This is a an operator with the main purpose of adding labels to
Nodes based on their names, immediately when Nodes are created. This is done
by using a mutating admission webhook.

This functionality is needed in case you can't use kubelet to add the labels
directly, and no other component should see the Node without those labels.

For being able to maintain the labels after node creation as well, the operator
will also add and optionally modify and delete labels on existing nodes in case
the configuration changes.

## Deployment

Until this available in the Openshift OperatorHub, the easiest way to install
is:

`operator-sdk run bundle quay.io/openshift-kni/node-label-operator-bundle:latest`

> Note: The `latest` tag is unstable, you might want to ask for the latest 
> stable version to use.

## Configuration

The operator has two CustomResourceDefinitions (CRDs) for the configuration
of labels:

### The Labels CRD

```go
// LabelsSpec defines the desired state of Labels
type LabelsSpec struct {
	// Rules defines a list of rules
	Rules []Rule `json:"rules"`
}

type Rule struct {
	// NodeNames defines a list of node name patterns for which the given labels should be set. 
	//String start and end anchors (^/$) will be added automatically 
	NodeNamePatterns []string `json:"nodeNamePatterns"`

	// Label defines the labels which should be set if one of the node name patterns matches
	// Format of label must be domain/name=value
	Labels []string `json:"labels"`
}
```

Creating instances of his CRD defines which labels should be added to which
nodes. A node can match with multiple CRs to accumulate multiple sets of labels.

### The OwnedLabels CRD

```go
// OwnedLabelsSpec defines the desired state of OwnedLabels
type OwnedLabelsSpec struct {
	// Domain defines the label domain which is owned by this operator
	// If a node label
	// - matches this domain AND
	// - matches the namePattern if given AND
	// - no label rule matches
	// then the label will be removed
	Domain *string `json:"domain,omitempty"`

	// NamePattern defines the label name pattern which is owned by this operator
	// If a node label
	// - matches this name pattern AND
	// - matches the domain if given AND
	// - no label rule matches
	// then the label will be removed
	// String start and end anchors (^/$) will be added automatically
	NamePattern *string `json:"namePattern,omitempty"`
}
```

Creating instances of this CRD defines which labels are "owned" by the operator.
Owned labels will be deleted in case no label rule matches anymore. Otherwise
the operator will only add labels or update label *values*.

### Example

Consider deployment of these manifests:

```yaml
apiVersion: node-labels.openshift.io/v1beta1
kind: Labels
metadata:
  name: labels-sample1
spec:
  rules:
    - nodeNamePatterns:
        - worker-0
      labels:
        - test.openshift.io/foo1=bar1
        - example.openshift.io/foo2=bar2
        - test.openshift.io/foo3=bar3
```
```yaml
apiVersion: node-labels.openshift.io/v1beta1
kind: Labels
metadata:
  name: labels-sample2
spec:
  rules:
    - nodeNamePatterns:
        - worker-0.*
      labels:
        - test.openshift.io/fooOther=barOther
```
```yaml
apiVersion: node-labels.openshift.io/v1beta1
kind: Labels
metadata:
  name: labels-sample3
spec:
  rules:
    - nodeNamePatterns:
        - dummy
      labels:
        - test.openshift.io/fooDummy=barDummy
```

```yaml
apiVersion: node-labels.openshift.io/v1beta1
kind: OwnedLabels
metadata:
  name: ownedlabels-sample
spec:
  domain: test.openshift.io
```

| Action | Result |
| -------| ------ |
| create node `worker-0` | The node will get labels `test.openshift.io/foo1=bar1`, `example.openshift.io/foo2=bar2`, `test.openshift.io/fooOther=barOther`, because the name matches the pattern of sample 1 and 2
| create node `worker-0-dummy` | The node will get label `test.openshift.io/fooOther=barOther`, because the name matches only the pattern of sample 2
| modify sample 1: change value `bar1` to `newBar1` | The node label *value* will be updated
| modify sample 1: change name `foo1` to `newFoo1` | The `foo1` node label will deleted and `newFoo1` created
| modify sample 1: change value `bar2` to `newBar2` | The node label *value* will be updated
| modify sample 1: change name `foo2` to `newFoo2` | Attention: a new label `example.openshift.io/newFoo2=bar2` will be added, but `example.openshift.io/foo2=newBar2` will stay unmodified! This is because the `example.openshift.io` domain is not owned by the operator (see last manifest), so existing label *names* will not be deleted / updated.
| modify sample 1: delete `newFoo1` label | The node label will be deleted
| modify sample 1: delete `newFoo2` label | The node label will NOT be deleted (see reason above)
| modify sample 1: modify `nodeNamePattern` to `worker-1` | The remaining node label of sample 1 `test.openshift.io/foo3=bar3` will be deleted from node `worker-0`
| | The sample 3 labels won't be applied to any node, because no node name matches

## License

Copyright 2021 Marc Sluiter

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.