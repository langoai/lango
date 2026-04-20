package exportability

func Evaluate(policy Policy, stage DecisionStage, refs []SourceRef) Receipt {
	lineage := make([]LineageSummary, 0, len(refs))
	if !policy.Enabled {
		return Receipt{
			Stage:       stage,
			State:       StateNeedsHumanReview,
			PolicyCode:  "review_policy_disabled",
			Explanation: "Exportability policy is disabled.",
			Lineage:     lineage,
		}
	}

	hasUserExportable := false
	for _, ref := range refs {
		rule := "source_class_ok"
		switch ref.Class {
		case "":
			rule = "metadata_missing"
		case ClassPrivateConfidential:
			rule = "highest_sensitivity_wins"
		case ClassUserExportable:
			hasUserExportable = true
		}
		lineage = append(lineage, LineageSummary{
			AssetID:    ref.AssetID,
			AssetLabel: ref.AssetLabel,
			Class:      ref.Class,
			Rule:       rule,
		})
	}

	for _, ref := range refs {
		if ref.Class == "" {
			return Receipt{
				Stage:       stage,
				State:       StateNeedsHumanReview,
				PolicyCode:  "review_metadata_conflict",
				Explanation: "Source metadata is incomplete or conflicting.",
				Lineage:     lineage,
			}
		}
		if ref.Class == ClassPrivateConfidential {
			return Receipt{
				Stage:       stage,
				State:       StateBlocked,
				PolicyCode:  "blocked_private_source",
				Explanation: "Artifact includes a private-confidential source.",
				Lineage:     lineage,
			}
		}
	}

	policyCode := "allowed_public_only"
	if hasUserExportable {
		policyCode = "allowed_user_exportable"
	}

	return Receipt{
		Stage:       stage,
		State:       StateExportable,
		PolicyCode:  policyCode,
		Explanation: "Artifact is exportable under source-based policy.",
		Lineage:     lineage,
	}
}
