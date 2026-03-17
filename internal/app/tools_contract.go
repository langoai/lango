package app

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/toolparam"
)

// buildContractTools creates agent tools for smart contract interaction.
func buildContractTools(caller *contract.Caller) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "contract_read",
			Description: "Read data from a smart contract (view/pure call, no gas cost)",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"address": map[string]interface{}{"type": "string", "description": "Contract address (0x...)"},
					"abi":     map[string]interface{}{"type": "string", "description": "Contract ABI as JSON string"},
					"method":  map[string]interface{}{"type": "string", "description": "Method name to call"},
					"args": map[string]interface{}{
						"type":        "array",
						"description": "Method arguments (optional)",
						"items":       map[string]interface{}{"type": "string"},
					},
					"chainId": map[string]interface{}{"type": "integer", "description": "Chain ID (optional, uses configured default)"},
				},
				"required": []string{"address", "abi", "method"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				req, err := parseContractCallParams(params)
				if err != nil {
					return nil, err
				}
				result, err := caller.Read(ctx, *req)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"data": result.Data,
				}, nil
			},
		},
		{
			Name:        "contract_call",
			Description: "Send a state-changing transaction to a smart contract (costs gas, may transfer ETH)",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"address": map[string]interface{}{"type": "string", "description": "Contract address (0x...)"},
					"abi":     map[string]interface{}{"type": "string", "description": "Contract ABI as JSON string"},
					"method":  map[string]interface{}{"type": "string", "description": "Method name to call"},
					"args": map[string]interface{}{
						"type":        "array",
						"description": "Method arguments (optional)",
						"items":       map[string]interface{}{"type": "string"},
					},
					"value":   map[string]interface{}{"type": "string", "description": "ETH value to send (e.g. '0.01'), optional"},
					"chainId": map[string]interface{}{"type": "integer", "description": "Chain ID (optional, uses configured default)"},
				},
				"required": []string{"address", "abi", "method"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				req, err := parseContractCallParams(params)
				if err != nil {
					return nil, err
				}
				// Parse optional ETH value (in wei or decimal ETH).
				if valStr, ok := params["value"].(string); ok && valStr != "" {
					ethWei, parseErr := parseETHValue(valStr)
					if parseErr != nil {
						return nil, fmt.Errorf("parse value %q: %w", valStr, parseErr)
					}
					req.Value = ethWei
				}
				result, err := caller.Write(ctx, *req)
				if err != nil {
					return nil, err
				}
				resp := map[string]interface{}{
					"txHash": result.TxHash,
				}
				if result.GasUsed > 0 {
					resp["gasUsed"] = result.GasUsed
				}
				return resp, nil
			},
		},
		{
			Name:        "contract_abi_load",
			Description: "Pre-load and cache a contract ABI for faster subsequent calls",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"address": map[string]interface{}{"type": "string", "description": "Contract address (0x...)"},
					"abi":     map[string]interface{}{"type": "string", "description": "Contract ABI as JSON string"},
					"chainId": map[string]interface{}{"type": "integer", "description": "Chain ID (optional, uses configured default)"},
				},
				"required": []string{"address", "abi"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				addrStr, err := toolparam.RequireString(params, "address")
				if err != nil {
					return nil, err
				}
				abiJSON, err := toolparam.RequireString(params, "abi")
				if err != nil {
					return nil, err
				}
				chainID := int64(toolparam.OptionalInt(params, "chainId", 0))
				addr := common.HexToAddress(addrStr)
				if err := caller.LoadABI(chainID, addr, abiJSON); err != nil {
					return nil, err
				}
				return map[string]interface{}{
					"status":  "loaded",
					"address": addr.Hex(),
				}, nil
			},
		},
	}
}

// parseContractCallParams extracts a ContractCallRequest from tool parameters.
func parseContractCallParams(params map[string]interface{}) (*contract.ContractCallRequest, error) {
	addrStr, err := toolparam.RequireString(params, "address")
	if err != nil {
		return nil, err
	}
	abiJSON, err := toolparam.RequireString(params, "abi")
	if err != nil {
		return nil, err
	}
	method, err := toolparam.RequireString(params, "method")
	if err != nil {
		return nil, err
	}

	chainID := int64(toolparam.OptionalInt(params, "chainId", 0))

	var args []interface{}
	if rawArgs, ok := params["args"].([]interface{}); ok {
		args = rawArgs
	}

	return &contract.ContractCallRequest{
		ChainID: chainID,
		Address: common.HexToAddress(addrStr),
		ABI:     abiJSON,
		Method:  method,
		Args:    args,
	}, nil
}

// parseETHValue converts a decimal ETH string (e.g. "0.01") to wei.
func parseETHValue(s string) (*big.Int, error) {
	rat := new(big.Rat)
	if _, ok := rat.SetString(s); !ok {
		return nil, fmt.Errorf("invalid ETH amount: %q", s)
	}
	// 1 ETH = 10^18 wei
	weiPerETH := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	rat.Mul(rat, new(big.Rat).SetInt(weiPerETH))
	if !rat.IsInt() {
		return nil, fmt.Errorf("ETH amount %q has too many decimal places", s)
	}
	return rat.Num(), nil
}
