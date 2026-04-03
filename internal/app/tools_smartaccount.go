package app

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/agent"
	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/toolparam"
	"github.com/langoai/lango/internal/smartaccount/paymaster"
	"github.com/langoai/lango/internal/wallet"
)

// buildSmartAccountTools creates the agent tools for the smart account subsystem.
func buildSmartAccountTools(sac *smartAccountComponents) []*agent.Tool {
	tools := []*agent.Tool{
		smartAccountDeployTool(sac),
		smartAccountInfoTool(sac),
		sessionKeyCreateTool(sac),
		sessionKeyListTool(sac),
		sessionKeyRevokeTool(sac),
		sessionExecuteTool(sac),
		policyCheckTool(sac),
		moduleInstallTool(sac),
		moduleUninstallTool(sac),
		spendingStatusTool(sac),
		paymasterStatusTool(sac),
		paymasterApproveTool(sac),
	}
	return tools
}

func smartAccountDeployTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "smart_account_deploy",
		Description: "Deploy a new Safe smart account with ERC-7579 modules",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			info, err := sac.manager.GetOrDeploy(ctx)
			if err != nil {
				return nil, fmt.Errorf("deploy smart account: %w", err)
			}
			modules := make([]map[string]interface{}, len(info.Modules))
			for i, m := range info.Modules {
				modules[i] = map[string]interface{}{
					"address": m.Address.Hex(),
					"type":    m.Type.String(),
					"name":    m.Name,
				}
			}
			return map[string]interface{}{
				"address":      info.Address.Hex(),
				"isDeployed":   info.IsDeployed,
				"ownerAddress": info.OwnerAddress.Hex(),
				"chainId":      info.ChainID,
				"entryPoint":   info.EntryPoint.Hex(),
				"modules":      modules,
			}, nil
		},
	}
}

func smartAccountInfoTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "smart_account_info",
		Description: "Get smart account information without deploying",
		SafetyLevel: agent.SafetyLevelSafe,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			info, err := sac.manager.Info(ctx)
			if err != nil {
				return nil, fmt.Errorf("get smart account info: %w", err)
			}
			modules := make([]map[string]interface{}, len(info.Modules))
			for i, m := range info.Modules {
				modules[i] = map[string]interface{}{
					"address": m.Address.Hex(),
					"type":    m.Type.String(),
					"name":    m.Name,
				}
			}
			return map[string]interface{}{
				"address":      info.Address.Hex(),
				"isDeployed":   info.IsDeployed,
				"ownerAddress": info.OwnerAddress.Hex(),
				"chainId":      info.ChainID,
				"entryPoint":   info.EntryPoint.Hex(),
				"modules":      modules,
			}, nil
		},
	}
}

func sessionKeyCreateTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "session_key_create",
		Description: "Create a new session key with scoped permissions (targets, functions, spend limit, duration)",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"targets": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Allowed target contract addresses (hex)",
				},
				"functions": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Allowed function selectors (4-byte hex, e.g. '0xa9059cbb')",
				},
				"spend_limit": map[string]interface{}{
					"type":        "string",
					"description": "Maximum spend in USDC (e.g. '10.00')",
				},
				"duration": map[string]interface{}{
					"type":        "string",
					"description": "Session duration (e.g. '1h', '30m', '24h')",
				},
				"parent_id": map[string]interface{}{
					"type":        "string",
					"description": "Parent session ID for task-scoped child sessions (optional)",
				},
			},
			"required": []string{"targets", "duration"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Parse targets
			rawTargets, _ := params["targets"].([]interface{})
			targets := make([]common.Address, 0, len(rawTargets))
			for _, rt := range rawTargets {
				if s, ok := rt.(string); ok {
					targets = append(targets, common.HexToAddress(s))
				}
			}

			// Parse functions
			var functions []string
			if rawFns, ok := params["functions"].([]interface{}); ok {
				for _, rf := range rawFns {
					if s, ok := rf.(string); ok {
						functions = append(functions, s)
					}
				}
			}

			// Parse spend limit
			var spendLimit *big.Int
			if limitStr, ok := params["spend_limit"].(string); ok && limitStr != "" {
				parsed, err := wallet.ParseUSDC(limitStr)
				if err != nil {
					return nil, fmt.Errorf("parse spend_limit %q: %w", limitStr, err)
				}
				spendLimit = parsed
			}

			// Parse duration
			durationStr, _ := params["duration"].(string)
			if durationStr == "" {
				durationStr = "1h"
			}
			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				return nil, fmt.Errorf("parse duration %q: %w", durationStr, err)
			}

			parentID, _ := params["parent_id"].(string)

			now := time.Now()
			pol := sa.SessionPolicy{
				AllowedTargets:   targets,
				AllowedFunctions: functions,
				SpendLimit:       spendLimit,
				ValidAfter:       now,
				ValidUntil:       now.Add(duration),
			}

			sk, err := sac.sessionManager.Create(ctx, pol, parentID)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"sessionId": sk.ID,
				"address":   sk.Address.Hex(),
				"expiresAt": sk.ExpiresAt.Format(time.RFC3339),
				"parentId":  sk.ParentID,
				"targets":   len(sk.Policy.AllowedTargets),
				"functions": len(sk.Policy.AllowedFunctions),
			}, nil
		},
	}
}

func sessionKeyListTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "session_key_list",
		Description: "List all session keys and their status",
		SafetyLevel: agent.SafetyLevelSafe,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			keys, err := sac.sessionManager.List(ctx)
			if err != nil {
				return nil, err
			}
			result := make([]map[string]interface{}, len(keys))
			for i, sk := range keys {
				status := "active"
				if sk.Revoked {
					status = "revoked"
				} else if sk.IsExpired() {
					status = "expired"
				}
				result[i] = map[string]interface{}{
					"sessionId": sk.ID,
					"address":   sk.Address.Hex(),
					"status":    status,
					"parentId":  sk.ParentID,
					"expiresAt": sk.ExpiresAt.Format(time.RFC3339),
					"createdAt": sk.CreatedAt.Format(time.RFC3339),
				}
			}
			return map[string]interface{}{
				"sessions": result,
				"total":    len(result),
			}, nil
		},
	}
}

func sessionKeyRevokeTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "session_key_revoke",
		Description: "Revoke a session key and all its child sessions",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "Session key ID to revoke",
				},
			},
			"required": []string{"session_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			sessionID, err := toolparam.RequireString(params, "session_id")
			if err != nil {
				return nil, err
			}
			if err := sac.sessionManager.Revoke(ctx, sessionID); err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"sessionId": sessionID,
				"status":    "revoked",
			}, nil
		},
	}
}

func sessionExecuteTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "session_execute",
		Description: "Execute a contract call using a session key (signs with session key, submits via bundler)",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "Session key ID to use for signing",
				},
				"target": map[string]interface{}{
					"type":        "string",
					"description": "Target contract address (hex)",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "ETH value to send in wei (default '0')",
				},
				"data": map[string]interface{}{
					"type":        "string",
					"description": "Call data in hex (e.g. '0xa9059cbb...')",
				},
				"function_sig": map[string]interface{}{
					"type":        "string",
					"description": "Function signature for policy tracking (e.g. 'transfer(address,uint256)')",
				},
			},
			"required": []string{"session_id", "target"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			sessionID, err := toolparam.RequireString(params, "session_id")
			if err != nil {
				return nil, err
			}
			targetStr, err := toolparam.RequireString(params, "target")
			if err != nil {
				return nil, err
			}

			target := common.HexToAddress(targetStr)

			// Parse value
			value := new(big.Int)
			if valStr, ok := params["value"].(string); ok && valStr != "" {
				parsed, ok := new(big.Int).SetString(valStr, 10)
				if !ok {
					return nil, fmt.Errorf("parse value %q: invalid integer", valStr)
				}
				value = parsed
			}

			// Parse call data
			var callData []byte
			if dataStr, ok := params["data"].(string); ok && dataStr != "" {
				dataStr = strings.TrimPrefix(dataStr, "0x")
				decoded, err := hex.DecodeString(dataStr)
				if err != nil {
					return nil, fmt.Errorf("decode data hex: %w", err)
				}
				callData = decoded
			}

			functionSig, _ := params["function_sig"].(string)

			// Build the contract call
			call := sa.ContractCall{
				Target:      target,
				Value:       value,
				Data:        callData,
				FunctionSig: functionSig,
			}

			// Validate against policy engine
			if sac.policyEngine != nil {
				if err := sac.policyEngine.Validate(target, &call); err != nil {
					return nil, fmt.Errorf("policy check: %w", err)
				}
			}

			// Sign the UserOp with the session key
			stubOp := &sa.UserOperation{
				Sender:               target,
				Nonce:                big.NewInt(0),
				InitCode:             []byte{},
				CallData:             callData,
				CallGasLimit:         big.NewInt(0),
				VerificationGasLimit: big.NewInt(0),
				PreVerificationGas:   big.NewInt(0),
				MaxFeePerGas:         big.NewInt(0),
				MaxPriorityFeePerGas: big.NewInt(0),
				PaymasterAndData:     []byte{},
				Signature:            []byte{},
			}
			if _, err := sac.sessionManager.SignUserOp(ctx, sessionID, stubOp); err != nil {
				return nil, fmt.Errorf("sign with session key: %w", err)
			}

			// Execute via the account manager
			txHash, execErr := sac.manager.Execute(ctx, []sa.ContractCall{call})
			if execErr != nil {
				return nil, fmt.Errorf("execute call: %w", execErr)
			}

			// Record spend if value > 0
			if value.Sign() > 0 && sac.onChainTracker != nil {
				sac.onChainTracker.Record(sessionID, value)
			}

			return map[string]interface{}{
				"txHash":    txHash,
				"sessionId": sessionID,
				"target":    target.Hex(),
			}, nil
		},
	}
}

func policyCheckTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "policy_check",
		Description: "Check if a contract call would pass the policy engine without executing it",
		SafetyLevel: agent.SafetyLevelSafe,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target": map[string]interface{}{
					"type":        "string",
					"description": "Target contract address (hex)",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "ETH value in wei (default '0')",
				},
				"function_sig": map[string]interface{}{
					"type":        "string",
					"description": "Function signature (e.g. 'transfer(address,uint256)')",
				},
			},
			"required": []string{"target"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			targetStr, err := toolparam.RequireString(params, "target")
			if err != nil {
				return nil, err
			}
			target := common.HexToAddress(targetStr)

			value := new(big.Int)
			if valStr, ok := params["value"].(string); ok && valStr != "" {
				parsed, ok := new(big.Int).SetString(valStr, 10)
				if ok {
					value = parsed
				}
			}

			functionSig, _ := params["function_sig"].(string)

			call := &sa.ContractCall{
				Target:      target,
				Value:       value,
				FunctionSig: functionSig,
			}

			if err := sac.policyEngine.Validate(target, call); err != nil {
				return map[string]interface{}{
					"allowed": false,
					"reason":  err.Error(),
				}, nil
			}
			return map[string]interface{}{
				"allowed": true,
				"target":  target.Hex(),
			}, nil
		},
	}
}

func moduleInstallTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "module_install",
		Description: "Install an ERC-7579 module on the smart account (validator, executor, hook, or fallback)",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"module_type": map[string]interface{}{
					"type":        "integer",
					"description": "Module type: 1=validator, 2=executor, 3=fallback, 4=hook",
					"enum":        []int{1, 2, 3, 4},
				},
				"address": map[string]interface{}{
					"type":        "string",
					"description": "Module contract address (hex)",
				},
				"init_data": map[string]interface{}{
					"type":        "string",
					"description": "Module initialization data in hex (optional)",
				},
			},
			"required": []string{"module_type", "address"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Parse module type — JSON numbers come as float64
			moduleTypeRaw := params["module_type"]
			var moduleType sa.ModuleType
			switch v := moduleTypeRaw.(type) {
			case float64:
				moduleType = sa.ModuleType(uint8(v))
			case int:
				moduleType = sa.ModuleType(uint8(v))
			default:
				return nil, fmt.Errorf("module_type must be an integer (1-4)")
			}

			addrStr, err := toolparam.RequireString(params, "address")
			if err != nil {
				return nil, err
			}
			addr := common.HexToAddress(addrStr)

			var initData []byte
			if dataStr, ok := params["init_data"].(string); ok && dataStr != "" {
				dataStr = strings.TrimPrefix(dataStr, "0x")
				decoded, err := hex.DecodeString(dataStr)
				if err != nil {
					return nil, fmt.Errorf("decode init_data hex: %w", err)
				}
				initData = decoded
			}

			txHash, err := sac.manager.InstallModule(ctx, moduleType, addr, initData)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"txHash":     txHash,
				"moduleType": moduleType.String(),
				"address":    addr.Hex(),
				"status":     "installed",
			}, nil
		},
	}
}

func moduleUninstallTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "module_uninstall",
		Description: "Uninstall an ERC-7579 module from the smart account",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"module_type": map[string]interface{}{
					"type":        "integer",
					"description": "Module type: 1=validator, 2=executor, 3=fallback, 4=hook",
					"enum":        []int{1, 2, 3, 4},
				},
				"address": map[string]interface{}{
					"type":        "string",
					"description": "Module contract address (hex)",
				},
			},
			"required": []string{"module_type", "address"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			moduleTypeRaw := params["module_type"]
			var moduleType sa.ModuleType
			switch v := moduleTypeRaw.(type) {
			case float64:
				moduleType = sa.ModuleType(uint8(v))
			case int:
				moduleType = sa.ModuleType(uint8(v))
			default:
				return nil, fmt.Errorf("module_type must be an integer (1-4)")
			}

			addrStr, err := toolparam.RequireString(params, "address")
			if err != nil {
				return nil, err
			}
			addr := common.HexToAddress(addrStr)

			txHash, err := sac.manager.UninstallModule(ctx, moduleType, addr, nil)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"txHash":     txHash,
				"moduleType": moduleType.String(),
				"address":    addr.Hex(),
				"status":     "uninstalled",
			}, nil
		},
	}
}

func spendingStatusTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "spending_status",
		Description: "View on-chain spending status and registered module information",
		SafetyLevel: agent.SafetyLevelSafe,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "Session ID to query spending for (optional, queries all if omitted)",
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			result := map[string]interface{}{}

			// On-chain tracker spending
			if sessionID, ok := params["session_id"].(string); ok && sessionID != "" {
				spent := sac.onChainTracker.GetSpent(sessionID)
				result["sessionId"] = sessionID
				result["onChainSpent"] = spent.String()
			}

			// Module registry info
			modules := sac.moduleRegistry.List()
			modList := make([]map[string]interface{}, len(modules))
			for i, m := range modules {
				modList[i] = map[string]interface{}{
					"name":    m.Name,
					"address": m.Address.Hex(),
					"type":    m.Type.String(),
					"version": m.Version,
				}
			}
			result["registeredModules"] = modList

			return result, nil
		},
	}
}

func paymasterStatusTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "paymaster_status",
		Description: "Check paymaster configuration and USDC approval status for gasless transactions",
		SafetyLevel: agent.SafetyLevelSafe,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			result := map[string]interface{}{
				"enabled": sac.paymasterProvider != nil,
			}

			if sac.paymasterProvider != nil {
				result["provider"] = sac.paymasterProvider.Type()
			} else {
				result["provider"] = "none"
			}

			return result, nil
		},
	}
}

func paymasterApproveTool(sac *smartAccountComponents) *agent.Tool {
	return &agent.Tool{
		Name:        "paymaster_approve",
		Description: "Approve USDC spending for the paymaster to enable gasless transactions",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category:             "smartaccount",
			Activity:             agent.ActivityManage,
			RequiredCapabilities: []string{"payment"},
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token_address": map[string]interface{}{
					"type":        "string",
					"description": "USDC token contract address (hex)",
				},
				"paymaster_address": map[string]interface{}{
					"type":        "string",
					"description": "Paymaster contract address (hex)",
				},
				"amount": map[string]interface{}{
					"type":        "string",
					"description": "USDC amount to approve (e.g. '1000.00'). Use 'max' for unlimited approval.",
				},
			},
			"required": []string{"token_address", "paymaster_address", "amount"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			tokenStr, err := toolparam.RequireString(params, "token_address")
			if err != nil {
				return nil, err
			}
			pmStr, err := toolparam.RequireString(params, "paymaster_address")
			if err != nil {
				return nil, err
			}
			amountStr, err := toolparam.RequireString(params, "amount")
			if err != nil {
				return nil, err
			}

			tokenAddr := common.HexToAddress(tokenStr)
			pmAddr := common.HexToAddress(pmStr)

			var amount *big.Int
			if amountStr == "max" {
				// MaxUint256
				amount = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
			} else {
				parsed, err := wallet.ParseUSDC(amountStr)
				if err != nil {
					return nil, fmt.Errorf("parse amount %q: %w", amountStr, err)
				}
				amount = parsed
			}

			approval := paymaster.NewApprovalCall(tokenAddr, pmAddr, amount)

			call := sa.ContractCall{
				Target:      approval.TokenAddress,
				Value:       big.NewInt(0),
				Data:        approval.ApproveCalldata,
				FunctionSig: "approve(address,uint256)",
			}

			txHash, err := sac.manager.Execute(ctx, []sa.ContractCall{call})
			if err != nil {
				return nil, fmt.Errorf("approve USDC: %w", err)
			}

			return map[string]interface{}{
				"txHash":    txHash,
				"token":     tokenAddr.Hex(),
				"paymaster": pmAddr.Hex(),
				"amount":    amountStr,
				"status":    "approved",
			}, nil
		},
	}
}
