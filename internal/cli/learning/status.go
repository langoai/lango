package learning

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show learning and knowledge system configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			errorCorrection := true
			if cfg.Agent.ErrorCorrectionEnabled != nil {
				errorCorrection = *cfg.Agent.ErrorCorrectionEnabled
			}

			type statusOutput struct {
				KnowledgeEnabled       bool    `json:"knowledge_enabled"`
				MaxContextPerLayer     int     `json:"max_context_per_layer"`
				AnalysisTurnThreshold  int     `json:"analysis_turn_threshold"`
				AnalysisTokenThreshold int     `json:"analysis_token_threshold"`
				ErrorCorrectionEnabled bool    `json:"error_correction_enabled"`
				ConfidenceThreshold    float64 `json:"auto_apply_confidence_threshold"`
				GraphEnabled           bool    `json:"graph_enabled"`
				GraphBackend           string  `json:"graph_backend,omitempty"`
				EmbeddingProvider      string  `json:"embedding_provider,omitempty"`
				EmbeddingModel         string  `json:"embedding_model,omitempty"`
				RAGEnabled             bool    `json:"rag_enabled"`
			}

			out := statusOutput{
				KnowledgeEnabled:       cfg.Knowledge.Enabled,
				MaxContextPerLayer:     cfg.Knowledge.MaxContextPerLayer,
				AnalysisTurnThreshold:  cfg.Knowledge.AnalysisTurnThreshold,
				AnalysisTokenThreshold: cfg.Knowledge.AnalysisTokenThreshold,
				ErrorCorrectionEnabled: errorCorrection,
				ConfidenceThreshold:    0.7,
				GraphEnabled:           cfg.Graph.Enabled,
				GraphBackend:           cfg.Graph.Backend,
				EmbeddingProvider:      cfg.Embedding.Provider,
				EmbeddingModel:         cfg.Embedding.Model,
				RAGEnabled:             cfg.Embedding.RAG.Enabled,
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Printf("Learning Status\n")
			fmt.Printf("  Knowledge Enabled:       %v\n", out.KnowledgeEnabled)
			fmt.Printf("  Error Correction:        %v\n", out.ErrorCorrectionEnabled)
			fmt.Printf("  Confidence Threshold:    %.1f\n", out.ConfidenceThreshold)
			fmt.Printf("  Max Context/Layer:       %d\n", out.MaxContextPerLayer)
			fmt.Printf("  Analysis Turn Threshold: %d\n", out.AnalysisTurnThreshold)
			fmt.Printf("  Analysis Token Threshold:%d\n", out.AnalysisTokenThreshold)
			fmt.Println()
			fmt.Printf("Graph Learning\n")
			fmt.Printf("  Graph Enabled:           %v\n", out.GraphEnabled)
			if out.GraphEnabled {
				fmt.Printf("  Graph Backend:           %s\n", out.GraphBackend)
			}
			fmt.Println()
			fmt.Printf("Embedding & RAG\n")
			fmt.Printf("  Embedding Provider:      %s\n", out.EmbeddingProvider)
			fmt.Printf("  Embedding Model:         %s\n", out.EmbeddingModel)
			fmt.Printf("  RAG Enabled:             %v\n", out.RAGEnabled)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
