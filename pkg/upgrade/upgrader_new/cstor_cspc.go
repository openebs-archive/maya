package upgrader

// CSPCPatch is the patch required to upgrade CSPC
type CSPCPatch struct {
	*ResourcePatch
	Namespace string
}

// CSPCPatchOptions ...
type CSPCPatchOptions func(*CSPCPatch)

// WithCSPCResorcePatch ...
func WithCSPCResorcePatch(r *ResourcePatch) CSPCPatchOptions {
	return func(obj *CSPCPatch) {
		obj.ResourcePatch = r
	}
}

// NewCSPCPatch ...
func NewCSPCPatch(opts ...CSPCPatchOptions) *CSPCPatch {
	obj := &CSPCPatch{}
	for _, o := range opts {
		o(obj)
	}
	return obj
}

// PreUpgrade ...
func (obj *CSPCPatch) PreUpgrade() error {
	return nil
}

// Init initializes all the fields of the CSPCPatch
func (obj *CSPCPatch) Init() error {
	obj.Namespace = obj.OpenebsNamespace
	return nil
}

// Upgrade execute the steps to upgrade CSPC
func (obj *CSPCPatch) Upgrade() error {
	err := obj.Init()
	if err != nil {
		return err
	}
	err = obj.PreUpgrade()
	if err != nil {
		return err
	}
	res := *obj.ResourcePatch
	cspiList := []string{"sparse-pool-1-msjd"}
	for _, cspiObobj := range cspiList {
		res.Name = cspiObobj
		depend := NewCSPIPatch(
			WithCSPIResorcePatch(&res),
		)
		err = depend.Upgrade()
		if err != nil {
			return err
		}
	}
	return nil
}
