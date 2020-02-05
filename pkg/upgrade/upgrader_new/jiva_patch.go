package upgrader

// import (
// 	"github.com/openebs/maya/pkg/upgrade/patch"
// )

// // JivaPatch is the patch required to upgrade jiva volume
// type JivaPatch struct {
// 	*ResourcePatch
// 	Namespace string
// 	Replica   *patch.Deployment
// 	Target    *patch.Deployment
// 	Service   *patch.Service
// }

// // JivaPatchOptions ...
// type JivaPatchOptions func(*JivaPatch)

// // WithJivaResorcePatch ...
// func WithJivaResorcePatch(r *ResourcePatch) JivaPatchOptions {
// 	return func(j *JivaPatch) {
// 		j.ResourcePatch = r
// 	}
// }

// // WithJivaReplica ...
// func WithJivaReplica(r *patch.Deployment) JivaPatchOptions {
// 	return func(j *JivaPatch) {
// 		j.Replica = r
// 	}
// }

// // WithJivaTarget ...
// func WithJivaTarget(t *patch.Deployment) JivaPatchOptions {
// 	return func(j *JivaPatch) {
// 		j.Target = t
// 	}
// }

// // WithJivaService ...
// func WithJivaService(s *patch.Service) JivaPatchOptions {
// 	return func(j *JivaPatch) {
// 		j.Service = s
// 	}
// }

// // NewJivaPatch ...
// func NewJivaPatch(opts ...JivaPatchOptions) *JivaPatch {
// 	j := &JivaPatch{}
// 	for _, o := range opts {
// 		o(j)
// 	}
// 	return j
// }

// // PreUpgrade ...
// func (j *JivaPatpackage upgraderch) PreUpgrade() error {
// 	err := j.Service.PreChecks(j.From, j.To)
// 	if err != nil {
// 		return err
// 	}
// 	err = j.Replica.PreChecks(j.From, j.To)
// 	if err != nil {
// 		return err
// 	}
// 	err = j.Target.PreChecks(j.From, j.To)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // ReplicaUpgrade ...
// func (j *JivaPatch) ReplicaUpgrade() error {
// 	err := j.Replica.Patch(j.From, j.To)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // TargetUpgrade ...
// func (j *JivaPatch) TargetUpgrade() error {
// 	err := j.Target.Patch(j.From, j.To)
// 	if err != nil {
// 		return err
// 	}
// 	err = j.Service.Patch(j.From, j.To)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // Upgrade execute the steps to upgrade jiva volume
// func (j *JivaPatch) Upgrade() error {
// 	err := j.PreUpgrade()
// 	if err != nil {
// 		return err
// 	}
// 	err = j.ReplicaUpgrade()
// 	if err != nil {
// 		return err
// 	}
// 	err = j.TargetUpgrade()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
