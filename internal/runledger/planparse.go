package runledger

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	// ErrInvalidPlanJSON is returned when the planner output is not valid JSON.
	ErrInvalidPlanJSON = fmt.Errorf("invalid plan JSON")
	// ErrPlanValidation is returned when the plan fails schema validation.
	ErrPlanValidation = fmt.Errorf("plan validation failed")
)

// jsonBlockRe matches ```json ... ``` fenced blocks.
var jsonBlockRe = regexp.MustCompile("(?s)```json\\s*\n?(.*?)```")

// ParsePlannerOutput extracts and deserializes the planner's JSON output.
// It accepts either a raw JSON object or a JSON block inside markdown fences.
func ParsePlannerOutput(raw string) (*PlannerOutput, error) {
	jsonStr := extractJSON(raw)
	if jsonStr == "" {
		return nil, fmt.Errorf("%w: no JSON block found", ErrInvalidPlanJSON)
	}

	var plan PlannerOutput
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPlanJSON, err)
	}
	return &plan, nil
}

// ValidatePlanSchema checks the parsed plan against structural and semantic rules.
func ValidatePlanSchema(plan *PlannerOutput, validAgents []string) error {
	var errs []string

	// Goal is required.
	if plan.Goal == "" {
		errs = append(errs, "goal is required")
	}

	// At least one step.
	if len(plan.Steps) == 0 {
		errs = append(errs, "at least one step is required")
	}

	// Build agent set for validation.
	agentSet := make(map[string]bool, len(validAgents))
	for _, a := range validAgents {
		agentSet[a] = true
	}

	// Check step ID uniqueness + required fields.
	stepIDs := make(map[string]bool, len(plan.Steps))
	for i, s := range plan.Steps {
		if s.ID == "" {
			errs = append(errs, fmt.Sprintf("step[%d]: id is required", i))
			continue
		}
		if stepIDs[s.ID] {
			errs = append(errs, fmt.Sprintf("step[%d]: duplicate id %q", i, s.ID))
		}
		stepIDs[s.ID] = true

		if s.Goal == "" {
			errs = append(errs, fmt.Sprintf("step[%d] %q: goal is required", i, s.ID))
		}
		if s.OwnerAgent == "" {
			errs = append(errs, fmt.Sprintf("step[%d] %q: owner_agent is required", i, s.ID))
		} else if len(validAgents) > 0 && !agentSet[s.OwnerAgent] {
			errs = append(errs, fmt.Sprintf("step[%d] %q: unknown agent %q", i, s.ID, s.OwnerAgent))
		}

		if err := validateValidatorType(s.Validator.Type); err != nil {
			errs = append(errs, fmt.Sprintf("step[%d] %q: %v", i, s.ID, err))
		}
	}

	// Check dependency references exist and detect cycles.
	for i, s := range plan.Steps {
		for _, dep := range s.DependsOn {
			if !stepIDs[dep] {
				errs = append(errs, fmt.Sprintf("step[%d] %q: depends_on references unknown step %q", i, s.ID, dep))
			}
		}
	}
	if err := detectCycle(plan.Steps); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate acceptance criteria.
	for i, ac := range plan.AcceptanceCriteria {
		if ac.Description == "" {
			errs = append(errs, fmt.Sprintf("acceptance_criteria[%d]: description is required", i))
		}
		if err := validateValidatorType(ac.Validator.Type); err != nil {
			errs = append(errs, fmt.Sprintf("acceptance_criteria[%d]: %v", i, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %s", ErrPlanValidation, strings.Join(errs, "; "))
	}
	return nil
}

// ConvertPlanToRunData converts a validated PlannerOutput into Steps and AcceptanceCriteria
// suitable for journal storage.
func ConvertPlanToRunData(plan *PlannerOutput) ([]Step, []AcceptanceCriterion) {
	steps := make([]Step, len(plan.Steps))
	for i, s := range plan.Steps {
		steps[i] = Step{
			StepID:      s.ID,
			Index:       i,
			Goal:        s.Goal,
			OwnerAgent:  s.OwnerAgent,
			Status:      StepStatusPending,
			Validator:   s.Validator,
			ToolProfile: s.ToolProfile,
			MaxRetries:  DefaultMaxRetries,
			DependsOn:   s.DependsOn,
		}
		// Auto-infer tool profile from validator type if not specified.
		if len(steps[i].ToolProfile) == 0 {
			steps[i].ToolProfile = inferToolProfile(s.Validator.Type)
		}
	}

	criteria := make([]AcceptanceCriterion, len(plan.AcceptanceCriteria))
	for i, ac := range plan.AcceptanceCriteria {
		criteria[i] = AcceptanceCriterion{
			Description: ac.Description,
			Validator:   ac.Validator,
		}
	}

	return steps, criteria
}

func extractJSON(raw string) string {
	// Try fenced block first.
	matches := jsonBlockRe.FindStringSubmatch(raw)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	// Try raw JSON.
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "{") {
		return trimmed
	}
	return ""
}

func validateValidatorType(vt ValidatorType) error {
	switch vt {
	case ValidatorBuildPass, ValidatorTestPass, ValidatorFileChanged,
		ValidatorArtifactExists, ValidatorCommandPass, ValidatorOrchestratorApproval:
		return nil
	case "":
		return fmt.Errorf("validator type is required")
	default:
		return fmt.Errorf("unknown validator type %q", vt)
	}
}

// detectCycle uses Kahn's algorithm to detect dependency cycles.
func detectCycle(steps []StepInput) error {
	inDegree := make(map[string]int, len(steps))
	adj := make(map[string][]string, len(steps))
	for _, s := range steps {
		if _, ok := inDegree[s.ID]; !ok {
			inDegree[s.ID] = 0
		}
		for _, dep := range s.DependsOn {
			adj[dep] = append(adj[dep], s.ID)
			inDegree[s.ID]++
		}
	}

	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	visited := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		visited++
		for _, next := range adj[node] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if visited < len(steps) {
		return fmt.Errorf("dependency cycle detected among steps")
	}
	return nil
}

func inferToolProfile(vt ValidatorType) []string {
	switch vt {
	case ValidatorBuildPass, ValidatorTestPass, ValidatorFileChanged:
		return []string{string(ToolProfileCoding)}
	case ValidatorOrchestratorApproval:
		return []string{string(ToolProfileSupervisor)}
	default:
		return []string{string(ToolProfileCoding)}
	}
}
