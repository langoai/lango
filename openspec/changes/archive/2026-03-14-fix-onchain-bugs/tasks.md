## 1. Receipt Status Validation

- [x] 1.1 Add ErrTxReverted and ErrReceiptTimeout sentinel errors to contract/caller.go
- [x] 1.2 Update Write() to return error on receipt timeout instead of nil
- [x] 1.3 Add receipt.Status check after waitForReceipt — return ErrTxReverted on revert
- [x] 1.4 Wrap waitForReceipt timeout error with ErrReceiptTimeout sentinel

## 2. Bundler Nonce Fix

- [x] 2.1 Replace eth_getTransactionCount with eth_call to EntryPoint.getNonce(address,uint192) in bundler/client.go
- [x] 2.2 ABI-encode getNonce calldata with selector 0x35567e1a
- [x] 2.3 Decode ABI-encoded uint256 result using hexutil.Decode + big.Int.SetBytes

## 3. Gas Fee Parameters

- [x] 3.1 Add GasFees struct to bundler/types.go
- [x] 3.2 Implement GetGasFees() in bundler/client.go with eth_maxPriorityFeePerGas and eth_getBlockByNumber
- [x] 3.3 Add fallback to default 1.5 gwei if eth_maxPriorityFeePerGas fails
- [x] 3.4 Wire GetGasFees() into manager.go submitUserOp() before gas estimation

## 4. IsDeployed Improvement

- [x] 4.1 Add *ethclient.Client field to Factory struct
- [x] 4.2 Update NewFactory() signature to accept *ethclient.Client parameter
- [x] 4.3 Replace isModuleInstalled-based IsDeployed with rpc.CodeAt() implementation
- [x] 4.4 Update wiring_smartaccount.go to pass rpcClient to NewFactory
- [x] 4.5 Update cli/smartaccount/deps.go to pass rpcClient to NewFactory

## 5. Deploy Verification

- [x] 5.1 Remove dead code result.Data extraction path in factory.go Deploy()
- [x] 5.2 Add post-deploy IsDeployed() verification in manager.go GetOrDeploy()

## 6. UserOp Signing Fix

- [x] 6.1 Change SignMessage to SignTransaction in manager.go submitUserOp()

## 7. Test Updates

- [x] 7.1 Update factory_test.go for new NewFactory signature and IsDeployed behavior
- [x] 7.2 Update manager_test.go mock bundler servers to handle eth_call, eth_maxPriorityFeePerGas, eth_getBlockByNumber
- [x] 7.3 Update integration_test.go mock bundler server for new RPC methods
- [x] 7.4 Verify all tests pass with go test ./internal/contract/... ./internal/smartaccount/...
