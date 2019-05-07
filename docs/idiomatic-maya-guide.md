# Idiomatic Maya Guide

## Introduction

Following are some of the explanations thats comes up when we search for the term "idiomatic". They are:

1. Using, containing, or denoting expressions that are natural to a native speaker
2. Appropriate to the style of art or music associated with a particular period, individual, or group.

I do not remember hearing anything about being idiomatic for other languages. I guess it has become synonymous with the advent of golang. Being idiomatic definitely means more than just good coding practices. It tries to provide a common ground for each member of the team to communicate in a way that is understood clearly. Obviously this is not solved by programming language, not even a high level programming language. IMO current day programming languages provide a bunch of dialects and lets its users (i.e. developers) choose the one they like the most. The fact that there is a choice, it hits the team hard later in the project's release lifecycle. The project code may not look as bad as sphagetti but no good either to understand fast and hence implement features, fixes, etc. faster.

## If Maya is Go then why not Idiomatic Go

Being idiomatic in Maya includes all the idiomatic pieces in Go and also takes into consideration writing code that understands and responds to Kubernetes a lot better. In addition, team at Maya has tried to put all their storage learnings so far into what is being termed as Idiomatic Maya where Maya is control plane for OpenEBS.

This document lists some of the guidelines and their corresponding examples that we are trying to follow in Maya to be idiomatic in true sense (there could be difference in opinions but we are trying to make sure that we all remain on the same page and hence it doesn't look confusing going further) -

## Naming Conventions

- Try to use names based on what the logic provides and not based on what the logic contains

```go
import (
  // Bad
  // alias `algorithm` conveys about what it contains i.e. some algorithm
  algorithm "github.com/openebs/maya/pkg/algorithm/nodeSelect/v1alpha1"
)
```

vs

```go
import (
  // Good
  // `nodeselect` tries to convey what the logic provides
  // this seems more natural way to express
  nodeselect "github.com/openebs/maya/pkg/algorithm/nodeSelect/v1alpha1"
)
```

```go
import(
// Bad
// upgrade looks more generic i.e upgrade package can have
// different other sub-packages like result which can be imported
upgrade "github.com/openebs/maya/pkg/upgrade/result/v1alpha1"
)
```

vs

```go
import(
// Good
// upgraderesult looks more specific and to the point i.e. the alias
// is for 'result' which is inside upgrade
upgraderesult "github.com/openebs/maya/pkg/upgrade/result/v1alpha1"
)
```

```go
// Bad
// taskPatch represents details for runtask patch operation
type taskPatch struct {}
```

vs

```go
// Good
// patch represents details for runtask patch operation
type patch struct {}
```

```go
// redundant (patch) -- not ok
patch.IsValidPatchType()
```

vs

```go
// Good
patch.IsValidType()
```

## Indentation (For enhancing readability)

We try to keep a line of code limited to not more than 80 characters so that the visibility and readability of the code looks better and aligned.

```go
// Bad
p, err := patch.BuilderForRuntask("UpgradeResult", m.runtask.Spec.Task, m.templateValues).AddCheck(patch.IsValidType()).Build()
```

vs

```go
// Good (looks more readable if aligned this way)
p, err := patch.
  BuilderForRuntask("UpgradeResult", m.runtask.Spec.Task, m.templateValues).
  AddCheck(patch.IsValidType()).
  Build()
```

```go
// Bad
p, err := upgraderesult.KubeClient(upgraderesult.WithNamespace(m.getRunTaskNamespace())).Patch(m.getTaskObjectName(), patch.Type, raw)
```

vs

```go
// Good (looks more readable)
p, err := upgraderesult.
  KubeClient(upgraderesult.WithNamespace(m.getRunTaskNamespace())).
  Patch(m.getTaskObjectName(), patch.Type, raw)
```

### How do you name your packages ?

```go
// Does below naming seem natural?
// Does full path reflect what the logic intends to provide?
// Does this adhere to golang's package naming convention?
//  i.e. nodeSelect vs nodeselect
//
// refer - https://blog.golang.org/package-names
// refer the other naming idioms mentioned in this doc

// Not Ok
pkg/upgrade/nodeSelect/v1alpha1
```

vs

```go
// Good (no camelcase)
pkg/upgrade/nodeselect/v1alpha1
```

## Do the names reflect their purpose ?

```go
// Not Ok

// assume a file pkg/upgrade/result/v1alpha1/kubernetes.go
// contains below code

// GetUpgradeResult returns an upgrade result instance
//
// Method name clashes with package name
//
// However, comment does help to some extent in understanding the purpose
// of this logic
func (k *kubeclient) GetUpgradeResult(name string, opts metav1.GetOptions) (*apis.UpgradeResult, error) {
    if strings.TrimSpace(name) == "" {
        return nil, errors.New("failed to get upgrade result: missing upgradeResult name")
    }
    cs, err := k.getClientOrCached()
    if err != nil {
        return nil, err
    }
    return k.get(cs, name, k.namespace, opts)
}
// Here the caller code will import this package as
upgraderesult "github.com/openebs/maya/pkg/upgrade/result/v1alpha1"

// And then the above method will be called as
upgraderesult.KubeClient().GetUpgradeResult(name,opts)
// The call above looks redundant since the word upgraderesult is being
// repeated.
```

vs

```go
// Good
// Get returns an upgrade result instance from kubernetes cluster
func (k *kubeclient) Get(name string, opts metav1.GetOptions) (*apis.UpgradeResult, error) {
    if strings.TrimSpace(name) == "" {
        return nil, errors.New("failed to get upgrade result: missing upgradeResult name")
    }
    cs, err := k.getClientOrCached()
    if err != nil {
        return nil, err
    }
    return k.get(cs, name, k.namespace, opts)
}
// Here the caller code will import this package as
upgraderesult "github.com/openebs/maya/pkg/upgrade/result/v1alpha1"

// And then the above method will be called as
upgraderesult.KubeClient().Get(name,opts)
// The call above looks more precise and clear since
// calling Get will return for package upgradeResult
// should return an upgrade result instance.
}
```

### Good names are great but avoid repeating them

```go
// Bad
type poolCreateConfig struct {
  // the word algorithm is repeated
  *algorithm.AlgorithmConfig
}
```

vs

```go
// Good
type poolCreateConfig struct {
  *algorithm.Config
}
```

```go
// Bad
poolconfig = &poolCreateConfig{
  // here the word algorithm gets repeated
  algorithm.NewAlgorithmConfig(spcGot),
}
```

vs

```go
// Good
poolconfig = &poolCreateConfig{
  algorithm.NewConfig(spcGot),
}
```

```go
// Bad
// the word new gets repeated; this is bad
pool, err := newClientSet.NewCasPool(spcGot, poolconfig)
```

vs

```go
// Good
p, err := cs.NewCasPool(spcGot, poolconfig)
```

## Make use of Builder Patterns

- Patterns are meant to be an effective communication technique between the developer as well as the reader of the code
- Once again remember natural way to express a specific thing (i.e. being idiomatic) helps us to achieve our overall objective

```go
// Without builder pattern
// patchUpgradeResult will patch an UpgradeResult as defined in the task
func (m *taskExecutor) patchUpgradeResult() (err error) {
    patch, err := asTaskPatch("UpgradeResult", m.runtask.Spec.Task, m.templateValues)
    if err != nil {
        return
    }
    pe, err := newTaskPatchExecutor(patch)
    if err != nil {
        return
    }
    raw, err := pe.toJson()
    if err != nil {
        return
    }
    // patch the upgrade result
    upgradeResult, err := m.getK8sClient().PatchUpgradeResult(m.getTaskObjectName(), m.getTaskRunNamespace(), pe.patchType(), raw)
    if err != nil {
        return
    }
    util.SetNestedField(m.templateValues, upgradeResult, string(v1alpha1.CurrentJSONResultTLP))
    return
}
```

vs

```go
// With Builder Pattern (recommended)
// patchUpgradeResult will patch an UpgradeResult as defined in the task
func (m *taskExecutor) patchUpgradeResult() (err error) {
    // build a runtask patch instance
    patch, err := patch.
        BuilderForRuntask("UpgradeResult", m.runtask.Spec.Task, m.templateValues).
        AddCheckf(patch.IsValidType(), "IsValidType").
        Build()
    if err != nil {
        return
    }
    // patch Upgrade Result
    p, err := upgraderesult.
        KubeClient(upgraderesult.WithNamespace(m.getTaskRunNamespace())).
        Patch(m.getTaskObjectName(), patch.Type, patch.Object)
    if err != nil {
        return
    }
    util.SetNestedField(m.templateValues, p, string(v1alpha1.CurrentJSONResultTLP))
    return
}
```

```go
// below does not express the builder pattern in its natural way
// Note: We can have difference of opinions, however glaring differences result into confusions
casPool, err := newClientSet.casPoolBuilder(casPool, spc, algorithmConfig)
```

vs

```go
b := caspool.Builder()
p, err := b.WithPool(casPool).
            WithClaim(spc).
            WithConfig(config).
            Build()
```

## Make use of Predicates (instead of if{}, else{}) for condition checks

- We are trying to make use of predicates instead of various blocks of if{}, else{}.

```go
// Predicate abstracts conditional logic w.r.t the patch instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*Patch) bool

// IsValidType returns true if provided patch
// type is one of the valid patch types
func (p *Patch) IsValidType() bool {
    return p.Type == types.JSONPatchType || p.Type == types.MergePatchType ||
        p.Type == types.StrategicMergePatchType
}

// Caller code can make use of these predicates in the following way :
// build a runtask patch instance
// checking if the patch type is valid or not
patch, err := patch.
    BuilderForRuntask("UpgradeResult", m.runtask.Spec.Task, m.templateValues).
    AddCheckf(patch.IsValidType(), "IsValidType").
    Build()
```

## Caller code should show the intent

```go
// The caller code should have the logic and capability to call methods as
// per requirement i.e. it should be provided with all the flexibilities
// that it requires i.e if it want to pass a namespace then it should
// be able to call some method such as WithNamespace() so that it can
// pass the namespace.

// An example caller code could look like -

// putUpgradeResult will put an upgrade result as defined in the task
func (m *taskExecutor) putUpgradeResult() (err error) {
    uresult, err := upgraderesult.
        BuilderForRuntask("UpgradeResult", m.runtask.Spec.Task, m.templateValues).
        Build()
    if err != nil {
        return
    }
    uraw, err := upgraderesult.
        KubeClient(upgraderesult.WithNamespace(m.getTaskRunNamespace())).
        CreateRaw(uresult)
    if err != nil {
        return
    }
    util.SetNestedField(m.templateValues, uraw, string(v1alpha1.CurrentJSONResultTLP))
    return
}
```

## Do we understand Table Driven Tests

```go
// There is no clarity on 
//  1. the inputs required for a test scenario & 
//  2. corresponding expectation after executing the test
//
// In other words, this does not look natural to Table Driven Tests
func TestNewCasPool(t *testing.T) {
	focs := &clientSet{
		oecs: openebsFakeClientset.NewSimpleClientset(),
	}
	focs.FakeDiskCreator()
	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	tests := map[string]struct {
		// fakestoragepoolclaim holds the fake storagepoolcalim object in test cases.
		fakestoragepoolclaim *apis.StoragePoolClaim
		autoProvisioning     bool
	}{
		// TestCase#1
		"SPC for manual provisioning with valid data": {
			autoProvisioning: false,
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
					Annotations: map[string]string{
						"cas.openebs.io/create-pool-template": "cstor-pool-create-default-0.7.0",
						"cas.openebs.io/delete-pool-template": "cstor-pool-delete-default-0.7.0",
					},
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: "disk",
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "striped",
					},
					Disks: apis.DiskAttr{
						DiskList: []string{"disk1", "disk2", "disk3"},
					},
				},
			},
		},
		"SPC for auto provisioning with valid data": {
			autoProvisioning: true,
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
					Annotations: map[string]string{
						"cas.openebs.io/create-pool-template": "cstor-pool-create-default-0.7.0",
						"cas.openebs.io/delete-pool-template": "cstor-pool-delete-default-0.7.0",
					},
				},
				Spec: apis.StoragePoolClaimSpec{
					MaxPools: 6,
					MinPools: 3,
					Type:     "disk",
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "mirrored",
					},
				},
			},
		},
	}
	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// newCasPool is the function under test.
			CasPool, err := focs.NewCasPool(test.fakestoragepoolclaim)
			fakeAlgoConf := fakeAlgorithmConfig(test.fakestoragepoolclaim)
			fakePoolConfig := &poolCreateConfig{
				fakeAlgoConf,
			}
			CasPool, err := focs.NewCasPool(test.fakestoragepoolclaim, fakePoolConfig)
			if err != nil || CasPool == nil {
				t.Errorf("Test case failed as expected nil error but error or CasPool object was nil:%s", name)
			}
		})
	}
}

// versus.
// below transforms above test logic into Table Driven format
//
// Query: Are you able to identify a table like structure?

func TestNewCasPool(t *testing.T) {
	tests := map[string]struct {
    isAuto    bool
    diskType  string
    poolType  string
    disks     []string
    poolCount int
    isErr     bool
	}{
	  // this is a tabular test format
	  // one can keep on adding more combinations in future
	  "t1": {false, "disk", "striped", []string{"d1", "d2", "d3"}, 3, false}
	  "t2": {false, "disk", "mirror", []string{"d1", "d2", "d3"}, 1, true}
	  "t3": {false, "disk", "raidz", []string{"d1", "d2", "d3"}, 0, true}
	  "t4": {true, "disk", "striped", []string{"d1", "d2", "d3"}, 3, false}
	  "t5": {true, "disk", "raidz1", []string{"d2", "d3"}, 3, false}
	  "t6": {false, "disk", "striped", []string{"d1", "d2", "d3"}, 3, false}
	  "t7": {false, "virtual", "mirror", []string{"d1", "d2", "d3"}, 3, true}
	  "t8": {false, "virtual", "striped", []string{"d1", "d2", "d3"}, 3, true}
	}
	for name, mock := range tests {
	  t.Run(name, func(t *testing.T){
	  	opts := []pool.BuildOption{
	      pool.WithDiskType(mock.diskType),
	      pool.WithPoolType(mock.poolType),
	      pool.WithPoolCount(mock.poolCount)
	    }
	    // Note: below is not a natural way to express test logic
	    // logic as shown below should be avoided as much as possible in
	    // unit tests; it leads to brittle test code
	    if mock.isAuto {
	      opts = append(opts, pool.Auto())
	    } else {
	      opts = append(opts, pool.WithDisks(mock.disks))
	    }
	    _, err := pool.New(opts)
	    if !mock.isErr && err != nil {
	      t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
	    }
	  })
	}
}

// versus.
//
// Above Table Driven test is divided into multiple Table Driven tests

func TestNewCasPoolAuto(t *testing.T) {
	tests := map[string]struct {
    diskType  string
    poolType  string
    poolCount int
    isErr     bool
	}{
	  // this is a tabular test format
	  // one can keep on adding more combinations in future
	  "t1": {"disk", "striped", 3, false}
	  "t2": {"disk", "mirror", 1, true}
	  "t3": {"disk", "raidz", 0, true}
	  "t4": {"virtual", "mirror", 3, true}
	  "t5": {"virtual", "striped", 3, true}
	}
	for name, mock := range tests {
	  t.Run(name, func(t *testing.T){
	    // below is more manageable way to express test logic; 
	    // it removes code brittleness as seen in the 1st & 2nd attempts
	    _, err := pool.New(
	          pool.Auto(), 
	          pool.WithDiskType(mock.diskType),
	          pool.WithPoolType(mock.poolType),
	          pool.WithPoolCount(mock.poolCount)
	    )
	    if !mock.isErr && err != nil {
	      t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
	    }
	  })
	}
}

// &&

func TestNewCasPoolManual(t *testing.T) {
	tests := map[string]struct {
    diskType  string
    poolType  string
    disks     []string
    poolCount int
    isErr     bool
	}{
	  // this is a tabular test format
	  // one can keep on adding more combinations in future
	  "t1": {"disk", "striped", []string{"d1", "d2", "d3"}, 3, false}
	  "t2": {"disk", "mirror", []string{"d1", "d2", "d3"}, 1, true}
	  "t3": {"disk", "raidz", []string{"d1", "d2", "d3"}, 0, true}
	  "t4": {"virtual", "mirror", []string{"d1", "d2", "d3"}, 3, true}
	  "t5": {"virtual", "striped", []string{"d1", "d2", "d3"}, 3, true}
	}
	for name, mock := range tests {
	  t.Run(name, func(t *testing.T){
	    // below is more manageable way to express test logic; 
	    // it removes code brittleness as seen in the 1st & 2nd attempts
	    _, err := pool.New(
	          pool.WithDiskType(mock.diskType),
	          pool.WithPoolType(mock.poolType),
	          pool.WithDisks(mock.disks),
	          pool.WithPoolCount(mock.poolCount)
	    )
	    _, err := pool.New(opts)
	    if !mock.isErr && err != nil {
	      t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
	    }
	  })
	}
}
```
