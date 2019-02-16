## Title: Openebs Operator specifications

### Authors:
- AmitKumarDas (github)
  - amitd (slack)
  - amit.das@mayadata.io (email)

### Reviewers:
- kmova (github)
- vishnuitta (github)
- sonasingh46 (github)

### Motivation:
This specifications is a simple to understand document that can be refered to during code reviews of openebs operator. One should be able to easily identity various points mentioned here with multiple usecases that it tries to implement. However, this is not a use-case document. It tries to manage the expectations of code reviewers and personas who want to understand/skim through the technical aspects minus any significant learning curve. In addition, this document tries **not** to be a deep dive technical journal & would like its readers to read/review the corresponding code instead.

In other words, I have tried to document the things I would have explained to someone reviewing the operator code. One can expect direct phrases like "It is XYZ", "It is **not** ABC", etc while reading through this document.

Note: This is intended to be a live as well versioned document
Note: Document avoids futuristic stuff and focuses on what is currently available or is currently being implemented
Note: Use cases related to openebs operator can be found in this document with section title _Issues_

### Abstract:
Openebs operator aims to be a helper to a human operator managing openebs. This operator manages the lifecycle of openebs on any Kubernetes cluster. It also acts as a handy tool that manages most of the non-functional aspects _(related to proper functioning)_ of openebs.

It will automate the process of _Install_, _Update_, _Upgrade_, _Probe / Checks_, _Self Healing_, and **more** with respect to openebs. It will work continuously to put the system into the desired state. A desired state is typically a list of desired specification(s) declared by this human operator to let openebs function at its prime.

### Status: Work In Progress (Monitor Component)
As part of helping the human operator, openebs operator should be able to accomplish following activities for proper functioning of openebs:

- Install -- _TODO_
- Reinstall i.e. Update with new configurations -- _TODO_
- Upgrade -- _TODO_
- Downgrade -- _TODO_
- Un-Install -- _TODO_
- **Monitor** -- _WIP_
- Self Heal -- _TODO_
- Config Management (includes validation) -- _TODO_

## Specifications:
### Provider
- Openebs operator considers container orchestrators as its provider
- Openebs controllers considers container orchestrators as its provider
- Openebs components considers container orchestrators as its provider
- Kubernetes is currently the only supported provider
- There are provisions at various levels which can be used to add other providers in future

### Operator
- It is a binary encapsulated in a docker image
- It consists of multiple controllers/reconcilers/watchers
- It is deployed as a kubernetes deployment
- It runs as a single replica deployment
- It manages openebs for a **single** kubernetes cluster
  - However, it is designed to manage openebs on a per namespace basis as well
  - NOTE: Namespace scoped feature is heavily dependent on individual workings of openebs component(s)
- Operator & controller logic makes use of kube-sigs' controller-runtime project
  - It refers to samples and examples provided in kube-sigs' kubebuilder project

### Openebs Component
- It is **NOT** a custom resource
- It tries to reflect an entire usecase
- For example, below is **single** component
  - a percona deployment
  - a persistent volume claim
  - a storage class
- For example, below represents maya api server as a **single** component
  - a maya api server deployment
  - a maya api server service
- For example, below represents openebs provisioner as a **single** component
  - an openebs provisioner deployment
- A component's template(s) is specified in a single _Catalog_
- A single component should be mapped to a single _Catalog_
- A component can represent one or more native kubernetes resource objects
- A component can represent one of more custom kubernetes resource objects

### Catalog
- Catalog is a kubernetes _Custom Resource_
- It provides one or more template(s) required to create a component
- It can be refered from one or more OpenebsCluster resources
- It can be watched/reconciled/controlled by **its controller**; 
  - Its own controller is known as _catalog controller_
  - This controller is part of **openebs operator** binary
- _Or_ it can watched/reconciled/controlled by **external controller**
  - One example of a catalog resource's external controller is OpenebsCluster controller
- It allows controllers to override the component properties such as:
  - _name_ of the component
  - _namespace_ of the component
  - _labels_ of the component
  - _annotations_ of the component
- It should not get deleted if atleast one component _(created via this catalog)_ is present in cluster
- Component that gets created by using a catalog's template should be identifiable
  - _openebs.io/catalog-reference_ can be set as an annotation in the component
- A component can be owned by a catalog resource:
  - if this catalog resource is controlled by _catalog controller_
  - component gets garbage collected if its owner gets deleted
- A component **may not be owned** by a catalog resource:
  - if this catalog resource is controlled by _OpenebsCluster controller_
  - component will be owned by OpenebsCluster resource
  - component will have a reference to catalog via its annotation _openebs.io/catalog-reference_
  - component is garbage collected if its owner gets deleted

### Catalog YAML

### OpenebsCluster
- It is a kubernetes Custom Resource
- It is observed/watched & controlled/reconciled by its controller
  - Its controller name is _OpenebsCluster controller_
  - This controller is part of **openebs operator** binary
- It does not manage openebs namespace
- It does not manage service accounts
- It does not manage openebs cluster role bindings
- On its creation related components are created
- On its deletion related components are deleted
- It refers to a catalog to create a component
- It owns the openebs component(s) that are specified in its specs
  - It owns the components it creates
- It adds/updates the properties of its owned components
- It errors out if catalog(s) refered to in its specs are not found
- It creates components if later are not found
- It reconciles the components if later are misconfigured
- It re-creates the components if later are deleted
- It deletes its owned yet duplicate components
  - a component which is owned by OpenebsCluster can get deplicated
  - if the component owned by openebscluster has its catalog reference changed
  - e.g. component catalog reference is changed from catalog named A to catalog named B
  - component related to catalog B get created
  - then the component related to catalog A is a duplicate
  - duplicate component refering to catalog A _(that is no longer used by openebs cluster)_ is deleted

### OpenebsCluster YAML

### KubeAssert
- It is a kubernetes _Custom Resource_
- It is watched/controlled/reconciled by its controller
  - Its controller is named as kubeassert controller
  - This controller is part of **openebs operator** binary
- It is an independent resource
  - It does not own any other kubernetes resource
- It is expected to get triggered periodically as it is managed as a kubernetes controller
- It can be created by a human operator
  - _Or_ it can be created by any tool
- It needs to be deleted by a human operator
  - _Or_ it can be deleted by any tool
- It can be thought of as a kubernetes cron job
  - However, it does not need an image unlike a cron job
- Assertions can be declaratively specified in this resource
- It can be used to monitor any kubernetes native resource object
- It can be used to monitor any kubernetes custom resource object
- It's status can be used to know the result of its assertion logic

### KubeAssert YAML

### Annotations
- These are few reference based annotations used for operator implementation
  - openebs.io/controller-reference
  - openebs.io/catalog-reference
  - openebs.io/hash-reference-list
  - openebs.io/casconfig-reference -- _FUTURE_

### Issues:
- https://github.com/openebs/openebs/issues/2343
- https://github.com/openebs/openebs/issues/2367

### Pull Requests:
- https://github.com/openebs/maya/pull/875
- https://github.com/openebs/maya/pull/900
- https://github.com/openebs/maya/pull/911
