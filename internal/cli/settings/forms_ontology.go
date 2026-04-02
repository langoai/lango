package settings

import (
	"fmt"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewOntologyForm creates the Ontology configuration form covering
// the main toggle, ACL, governance, and P2P exchange sub-sections.
func NewOntologyForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Ontology Configuration")

	enabled := &tuicore.Field{
		Key: "ontology_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Ontology.Enabled,
		Description: "Enable the ontology registry and seed migration",
	}
	form.AddField(enabled)
	isEnabled := func() bool { return enabled.Checked }

	// --- ACL ---
	aclEnabled := &tuicore.Field{
		Key: "ontology_acl_enabled", Label: "  ACL Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Ontology.ACL.Enabled,
		Description: "Enforce role-based access control on ontology operations",
		VisibleWhen: isEnabled,
	}
	form.AddField(aclEnabled)
	isACLEnabled := func() bool { return enabled.Checked && aclEnabled.Checked }

	form.AddField(&tuicore.Field{
		Key: "ontology_acl_roles", Label: "    Roles", Type: tuicore.InputText,
		Value:       formatKeyValuePairs(cfg.Ontology.ACL.Roles),
		Placeholder: "operator=write,librarian=read",
		Description: "Principal-to-permission mapping (read, write, admin)",
		VisibleWhen: isACLEnabled,
	})

	p2pPerm := cfg.Ontology.ACL.P2PPermission
	if p2pPerm == "" {
		p2pPerm = "write"
	}
	form.AddField(&tuicore.Field{
		Key: "ontology_acl_p2p_permission", Label: "    P2P Default Permission", Type: tuicore.InputSelect,
		Value:       p2pPerm,
		Options:     []string{"read", "write", "admin"},
		Description: "Default permission for peer: prefix principals",
		VisibleWhen: isACLEnabled,
	})

	// --- Governance ---
	govEnabled := &tuicore.Field{
		Key: "ontology_gov_enabled", Label: "  Governance Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Ontology.Governance.Enabled,
		Description: "Enforce schema lifecycle FSM on RegisterType/RegisterPredicate",
		VisibleWhen: isEnabled,
	}
	form.AddField(govEnabled)
	isGovEnabled := func() bool { return enabled.Checked && govEnabled.Checked }

	form.AddField(&tuicore.Field{
		Key: "ontology_gov_max_new_per_day", Label: "    Max New Per Day", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%d", cfg.Ontology.Governance.MaxNewPerDay),
		Placeholder: "20",
		Description: "Combined daily limit for type + predicate proposals",
		VisibleWhen: isGovEnabled,
	})
	form.AddField(&tuicore.Field{
		Key: "ontology_gov_quarantine_hrs", Label: "    Quarantine Period (hrs)", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%d", cfg.Ontology.Governance.QuarantinePeriodHrs),
		Placeholder: "24",
		Description: "Hours a proposal must wait before advancing",
		VisibleWhen: isGovEnabled,
	})
	form.AddField(&tuicore.Field{
		Key: "ontology_gov_shadow_hrs", Label: "    Shadow Mode Duration (hrs)", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%d", cfg.Ontology.Governance.ShadowModeDurationHrs),
		Placeholder: "168",
		Description: "Hours in shadow mode before eligible for promotion",
		VisibleWhen: isGovEnabled,
	})
	form.AddField(&tuicore.Field{
		Key: "ontology_gov_min_usage", Label: "    Min Usage for Promotion", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%d", cfg.Ontology.Governance.MinUsageForPromotion),
		Placeholder: "5",
		Description: "Minimum usage count for auto-promotion from shadow to active",
		VisibleWhen: isGovEnabled,
	})
	form.AddField(&tuicore.Field{
		Key: "ontology_gov_explosion_budget", Label: "    Schema Explosion Budget", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%d", cfg.Ontology.Governance.SchemaExplosionBudget),
		Placeholder: "100",
		Description: "Monthly limit for new proposals (type + predicate combined)",
		VisibleWhen: isGovEnabled,
	})

	// --- Exchange ---
	exEnabled := &tuicore.Field{
		Key: "ontology_ex_enabled", Label: "  Exchange Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Ontology.Exchange.Enabled,
		Description: "Enable P2P ontology exchange (requires P2P + Ontology)",
		VisibleWhen: isEnabled,
	}
	form.AddField(exEnabled)
	isExEnabled := func() bool { return enabled.Checked && exEnabled.Checked }

	form.AddField(&tuicore.Field{
		Key: "ontology_ex_min_trust_schema", Label: "    Min Trust for Schema", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.2f", cfg.Ontology.Exchange.MinTrustForSchema),
		Placeholder: "0.50",
		Description: "Minimum peer trust score for schema exchange",
		VisibleWhen: isExEnabled,
	})
	form.AddField(&tuicore.Field{
		Key: "ontology_ex_min_trust_facts", Label: "    Min Trust for Facts", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.2f", cfg.Ontology.Exchange.MinTrustForFacts),
		Placeholder: "0.70",
		Description: "Minimum peer trust score for fact exchange",
		VisibleWhen: isExEnabled,
	})

	autoImport := cfg.Ontology.Exchange.AutoImportMode
	if autoImport == "" {
		autoImport = "shadow"
	}
	form.AddField(&tuicore.Field{
		Key: "ontology_ex_auto_import_mode", Label: "    Auto-Import Mode", Type: tuicore.InputSelect,
		Value:       autoImport,
		Options:     []string{"shadow", "governed", "disabled"},
		Description: "How imported schemas enter the lifecycle: shadow, governed, or disabled",
		VisibleWhen: isExEnabled,
	})
	form.AddField(&tuicore.Field{
		Key: "ontology_ex_max_types", Label: "    Max Types Per Import", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%d", cfg.Ontology.Exchange.MaxTypesPerImport),
		Placeholder: "10",
		Description: "Maximum types accepted from a single peer exchange",
		VisibleWhen: isExEnabled,
	})

	return &form
}

