// This package deals with profiles. A profile in maya api service categorises
// the generic properties asociated with maya api service components (i.e. volume
// provisioners, orchestration providers).
//
// e.g. A profile consisting of a set of properties relevant to a persistent
// volume provisioner. This profile which is typically a set of key-value pairs
// can be applied with volume provisioner specific operations. Similarly, there
// can be a profile of properties that can be applied with orchestrator specific
// operations.
//
// Operator can create different variants of profile meant for persistent volume
// provisioners or for orchestration providers. The profiles can always be overridden
// with the properties specified at runtime.
package orchestrator
