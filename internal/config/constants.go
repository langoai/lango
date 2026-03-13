package config

// Validation maps for configuration values.
var (
	ValidLogLevels         = map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	ValidLogFormats        = map[string]bool{"json": true, "console": true}
	ValidSignerProviders   = map[string]bool{"local": true, "rpc": true, "enclave": true, "aws-kms": true, "gcp-kms": true, "azure-kv": true, "pkcs11": true}
	ValidWalletProviders   = map[string]bool{"local": true, "rpc": true, "composite": true}
	ValidZKPSchemes        = map[string]bool{"plonk": true, "groth16": true}
	ValidContainerRuntimes = map[string]bool{"auto": true, "docker": true, "gvisor": true, "native": true}
	ValidMCPTransports     = map[string]bool{"": true, "stdio": true, "http": true, "sse": true}
)
