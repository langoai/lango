package toolchain

import "sort"

// HookRegistry holds and runs pre/post hooks in priority order.
type HookRegistry struct {
	preHooks  []PreToolHook
	postHooks []PostToolHook
}

// NewHookRegistry creates a new HookRegistry ready for use.
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{}
}

// RegisterPre adds a pre-tool hook to the registry.
func (r *HookRegistry) RegisterPre(hook PreToolHook) {
	r.preHooks = append(r.preHooks, hook)
	sort.Slice(r.preHooks, func(i, j int) bool {
		return r.preHooks[i].Priority() < r.preHooks[j].Priority()
	})
}

// RegisterPost adds a post-tool hook to the registry.
func (r *HookRegistry) RegisterPost(hook PostToolHook) {
	r.postHooks = append(r.postHooks, hook)
	sort.Slice(r.postHooks, func(i, j int) bool {
		return r.postHooks[i].Priority() < r.postHooks[j].Priority()
	})
}

// PreHooks returns the registered pre-hooks (for diagnostics).
func (r *HookRegistry) PreHooks() []PreToolHook { return r.preHooks }

// PostHooks returns the registered post-hooks (for diagnostics).
func (r *HookRegistry) PostHooks() []PostToolHook { return r.postHooks }

// RunPre runs all pre-hooks in priority order.
// Returns the first Block result immediately.
// If multiple hooks return Modify, the last one's params win.
// Returns Continue with nil params if no hook blocks or modifies.
func (r *HookRegistry) RunPre(ctx HookContext) (PreHookResult, error) {
	result := PreHookResult{Action: Continue}
	for _, hook := range r.preHooks {
		hr, err := hook.Pre(ctx)
		if err != nil {
			return PreHookResult{}, err
		}
		switch hr.Action {
		case Block:
			return hr, nil
		case Modify:
			result = hr
			// Update params for subsequent hooks to see the modification.
			ctx.Params = hr.ModifiedParams
		}
	}
	return result, nil
}

// RunPost runs all post-hooks in priority order.
// Returns the first error encountered.
func (r *HookRegistry) RunPost(ctx HookContext, result interface{}, toolErr error) error {
	for _, hook := range r.postHooks {
		if err := hook.Post(ctx, result, toolErr); err != nil {
			return err
		}
	}
	return nil
}
