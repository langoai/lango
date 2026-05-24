package learning

import (
	"fmt"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show learning and knowledge system configuration",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
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

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), out)
			}

			writer := cmd.OutOrStdout()
			fmt.Fprintf(writer, "Learning Status\n")
			fmt.Fprintf(writer, "  Knowledge Enabled:       %v\n", out.KnowledgeEnabled)
			fmt.Fprintf(writer, "  Error Correction:        %v\n", out.ErrorCorrectionEnabled)
			fmt.Fprintf(writer, "  Confidence Threshold:    %.1f\n", out.ConfidenceThreshold)
			fmt.Fprintf(writer, "  Max Context/Layer:       %d\n", out.MaxContextPerLayer)
			fmt.Fprintf(writer, "  Analysis Turn Threshold: %d\n", out.AnalysisTurnThreshold)
			fmt.Fprintf(writer, "  Analysis Token Threshold:%d\n", out.AnalysisTokenThreshold)
			fmt.Fprintln(writer)
			fmt.Fprintf(writer, "Graph Learning\n")
			fmt.Fprintf(writer, "  Graph Enabled:           %v\n", out.GraphEnabled)
			if out.GraphEnabled {
				fmt.Fprintf(writer, "  Graph Backend:           %s\n", out.GraphBackend)
			}
			fmt.Fprintln(writer)
			fmt.Fprintf(writer, "Embedding & RAG\n")
			fmt.Fprintf(writer, "  Embedding Provider:      %s\n", out.EmbeddingProvider)
			fmt.Fprintf(writer, "  Embedding Model:         %s\n", out.EmbeddingModel)
			fmt.Fprintf(writer, "  RAG Enabled:             %v\n", out.RAGEnabled)

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
