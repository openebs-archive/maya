## Introduction
Following are some of the explanations thats comes up when we search for the term "idiomatic". They are:
1. using, containing, or denoting expressions that are natural to a native speaker
2. appropriate to the style of art or music associated with a particular period, individual, or group.

I do not remember hearing anything about being idiomatic for other languages. I guess it has become synonymous with the advent of golang. Being idiomatic definitely means more than just good coding practices. It tries to provide a common ground for each member of the team to communicate in a way that is understood clearly. Obviously this is not solved by programming language, not even a high level programming language. IMO current day programming languages provide a bunch of dialects and lets its users (i.e. developers) choose the one they like the most. The fact that there is a choice, it hits the team hard later in the project's release lifecycle. The project code may not look as bad as sphagetti but no good either to understand fast and hence implement features, fixes, etc. faster.

## If Maya is Go then why not Idiomatic Go
Being idiomatic in Maya includes all the idiomatic pieces in Go and also takes into consideration writing code that understands and responds to Kubernetes a lot better. In addition, team at Maya has tried to put all their storage learnings so far into what is being termed as Idiomatic Maya where Maya is control plane for OpenEBS.


## Devil is in the details
Below are some sample code reviews that describes with examples to write idiomatic maya code.

### Naming - Thumb Rule
- Use names based on what the logic provides and not based on what the logic contains

```go
import (
  // alias `algorithm` conveys about what it contains i.e. some algorithm
  algorithm "github.com/openebs/maya/pkg/algorithm/nodeSelect/v1alpha1"

  // vs.

  // `nodeselect` tries to convey what the logic provides
  // this seems more natural way to express
  nodeselect "github.com/openebs/maya/pkg/algorithm/nodeSelect/v1alpha1"
)
```

```go
nodeDisks, err := ac.NodeDiskSelector()
// `disks` is repeated; does not seem natural
if len(nodeDisks.Disks.Items) == 0 {
  // ...
}

// vs.

s, err := ac.NodeDiskSelector()
if len(s.Disks.Items) == 0 {
  // ...
}
```

### Do the names reflect their purpose?
```go
// assume a file pkg/algorithm/nodeSelect/v1alpha1/select_node.go 
// contains below code

// NodeDiskSelector selects a node and disks attached to it.
//
// Method name clashes with package name
// These are few questions that comes to mind:
//  1/ Does this logic return a node instance?
//  2/ Does this logic return a disk instance?
//
// However, comment does help to some extent in understanding the purpose 
// of this logic
func (ac *AlgorithmConfig) NodeDiskSelector() (*nodeDisk, error) {
	listDisk, err := ac.getDisk()
	if listDisk == nil || len(listDisk.Items) == 0 {
		return nil, errors.Wrapf(err, "no disk object found")
	}
	nodeDiskMap, err := ac.getCandidateNode(listDisk)
	if err != nil {
		return nil, err
	}
	selectedDisk := ac.selectNode(nodeDiskMap)
	return selectedDisk, nil
}

vs.

// change the file path to below
// pkg/nodeselect/v1alpha1/nodeselect.go

// NodeDiskSelector selects and returns appropriate node
//
// Note: It is assumed that the information required to select is 
// available in nodeselect instance
//
// Note: The return has to be an instance of node. This instance
// should be defined in this package
//
// Note: Below might indicate a stripped off version, however it is
// not. Code needs to be designed to place business logic in proper places.
// A good code reflects in its Unit Testing as well as in its caller code.
func (s *nodeselect) Get() (*node, error) {
	d, err := s.getDisks()
	if d == nil || len(d.Items) == 0 {
		return nil, errors.Wrapf(err, "no disks provided")
	}
	dl := disk.NewList(disk.WithNames(d.Items)).Filter(disk.IsFree)
	nl := node.NewList(node.WithDisks(dl.Item)).Filter(node.ContainsDisk)
	return nl.Filter(node.IsPoolFeasible())
}
```

### Good names are great but avoid repeating them

```go
type poolCreateConfig struct {
  // the word algorithm is repeated
	*algorithm.AlgorithmConfig
}

// vs.

type poolCreateConfig struct {
	*algorithm.Config
}
```

```go
poolconfig = &poolCreateConfig{
  // here the word algorithm gets repeated
	algorithm.NewAlgorithmConfig(spcGot),
}

// vs.

poolconfig = &poolCreateConfig{
	algorithm.NewConfig(spcGot),
}
```

```go
// the word new gets repeated; this is bad
pool, err := newClientSet.NewCasPool(spcGot, poolconfig)

// vs.

p, err := cs.NewCasPool(spcGot, poolconfig)
```

### How do you name your packages?

```go
// Does below naming seem natural?
// Does full path reflect what the logic intends to provide?
// Does this adhere to golang's package naming convention?
//  i.e. nodeSelect vs nodeselect
//
// refer - https://blog.golang.org/package-names
// refer the other naming idioms mentioned in this doc
pkg/algorithm/nodeSelect/v1alpha1

// vs.

pkg/nodeselect/v1alpha1
```

### Caller code should show the intent

```go
pool, err := newClientSet.NewCasPool(spcGot, poolconfig)

// vs.

// if caller needs a new instance of cas pool
p, err := caspool.New(...)

// vs. 
// if caller needs to fetch from some service
p, err := caspool.Get(..)
```

### Closed for modification yet open to extension
- Avoid changing function signature like the one shown below
- This can be caught by compilers, but is still a good one to sort during the design phase itself
- Think how readability suffers, if you need to add a few more arguments to below function

```diff
- func (newClientSet *clientSet) NewCasPool(spc *apis.StoragePoolClaim) (*apis.CasPool, error) {
+ func (newClientSet *clientSet) NewCasPool(spc *apis.StoragePoolClaim, algorithmConfig *poolCreateConfig) (*apis.CasPool, error) {
```

### Do we digress too much from original patterns : Builder Pattern
- Patterns are meant to be an effective communication technique between the developer as well as the reader of the code
- Once again remember natural way to express a specific thing (i.e. being idiomatic) helps us to achieve our overall objective

```go
// below does not express the builder pattern in its natural way
// Note: We can have difference of opinions, however glaring differences result into confusions
casPool, err := newClientSet.casPoolBuilder(casPool, spc, algorithmConfig)

// vs.
b := caspool.Builder()
p, err := b.WithPool(casPool).
            WithClaim(spc).
            WithConfig(config).
            Build()
```

### Do you identify boilerplate in your code (TODO)


### Do we understand Table Driven Tests

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

### Do you identify boilerplate in your test code (TODO)
- Understand what is boilerplate versus what you want to test
- This will help you in writing readable & hence maintainable test cases

### Enabling configurability at source code solves most of Unit Test's corner cases (TODO)
- Over-parameterize structs to allow tests to fine-tune their behavior
- It is okay to make these configurations unexported so only tests can set them


### go test has lot more to offer (TODO)
- go test as a tool is a fantastic taskrunner


### Have you ever considered **testing** as a public API (TODO)
- Practice of adopting testing.go or testing_*.go files
- These are exported APIs for the sole purpose of providing mocks, tests, harnesses, helpers, etc
- Allows other packages to test using above package's mocks without the need to write mocks
- Above naming convention allows one to easily find test helpers
