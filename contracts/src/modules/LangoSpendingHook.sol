// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title LangoSpendingHook — ERC-7579 TYPE_HOOK module for spending controls.
/// @notice Enforces per-transaction, daily, and cumulative spending limits per account/session key.
contract LangoSpendingHook {
    // ERC-7579 module type constants
    uint256 internal constant TYPE_HOOK = 4;
    uint256 internal constant DAY = 86400;

    struct SpendingConfig {
        uint256 perTxLimit;
        uint256 dailyLimit;
        uint256 cumulativeLimit;
        bool configured;
    }

    struct SpendState {
        uint256 dailySpent;
        uint256 dailyResetTimestamp;
        uint256 cumulativeSpent;
    }

    // account => spending configuration
    mapping(address => SpendingConfig) public configs;

    // account => session key => spend state
    mapping(address => mapping(address => SpendState)) public spendStates;

    // account => global spend state (address(0) as key)
    mapping(address => SpendState) public globalStates;

    event LimitsUpdated(address indexed account, uint256 perTxLimit, uint256 dailyLimit, uint256 cumulativeLimit);
    event SpendRecorded(address indexed account, address indexed sessionKey, uint256 amount);

    // ---- IERC7579Module ----

    /// @notice Called when this module is installed on an account.
    /// @param data ABI-encoded SpendingConfig (perTxLimit, dailyLimit, cumulativeLimit).
    function onInstall(bytes calldata data) external {
        if (data.length > 0) {
            (uint256 perTx, uint256 daily, uint256 cumulative) = abi.decode(data, (uint256, uint256, uint256));
            configs[msg.sender] = SpendingConfig({
                perTxLimit: perTx,
                dailyLimit: daily,
                cumulativeLimit: cumulative,
                configured: true
            });
            emit LimitsUpdated(msg.sender, perTx, daily, cumulative);
        }
    }

    /// @notice Called when this module is uninstalled.
    function onUninstall(bytes calldata) external {
        delete configs[msg.sender];
    }

    /// @notice Returns true if moduleTypeId == 4 (HOOK).
    function isModuleType(uint256 moduleTypeId) external pure returns (bool) {
        return moduleTypeId == TYPE_HOOK;
    }

    // ---- Hook ----

    /// @notice Pre-execution check. Validates spending limits.
    /// @param msgSender The original msg.sender (session key or account owner).
    /// @param value The ETH value being sent.
    /// @param msgData The call data (unused in current implementation).
    /// @return hookData Encoded context passed to postCheck.
    function preCheck(address msgSender, uint256 value, bytes calldata msgData)
        external
        returns (bytes memory hookData)
    {
        // Silence unused parameter warning
        msgData;

        SpendingConfig storage cfg = configs[msg.sender];
        if (!cfg.configured) {
            return abi.encode(msgSender, value);
        }

        // Per-transaction limit
        require(cfg.perTxLimit == 0 || value <= cfg.perTxLimit, "Hook: exceeds per-tx limit");

        // Update and check session key spend state
        SpendState storage state = spendStates[msg.sender][msgSender];
        _resetDailyIfNeeded(state);

        if (cfg.dailyLimit > 0) {
            require(state.dailySpent + value <= cfg.dailyLimit, "Hook: exceeds daily limit");
        }
        if (cfg.cumulativeLimit > 0) {
            require(state.cumulativeSpent + value <= cfg.cumulativeLimit, "Hook: exceeds cumulative limit");
        }

        // Update global state
        SpendState storage global = globalStates[msg.sender];
        _resetDailyIfNeeded(global);

        if (cfg.dailyLimit > 0) {
            require(global.dailySpent + value <= cfg.dailyLimit, "Hook: exceeds global daily limit");
        }
        if (cfg.cumulativeLimit > 0) {
            require(global.cumulativeSpent + value <= cfg.cumulativeLimit, "Hook: exceeds global cumulative limit");
        }

        // Record spend
        state.dailySpent += value;
        state.cumulativeSpent += value;
        global.dailySpent += value;
        global.cumulativeSpent += value;

        emit SpendRecorded(msg.sender, msgSender, value);

        return abi.encode(msgSender, value);
    }

    /// @notice Post-execution check. Currently a no-op.
    /// @param hookData Data returned from preCheck.
    function postCheck(bytes calldata hookData) external pure {
        // No-op: reserved for future post-execution validation.
        hookData;
    }

    // ---- Owner functions ----

    /// @notice Set or update spending limits for the calling account.
    function setLimits(uint256 perTxLimit, uint256 dailyLimit, uint256 cumulativeLimit) external {
        configs[msg.sender] = SpendingConfig({
            perTxLimit: perTxLimit,
            dailyLimit: dailyLimit,
            cumulativeLimit: cumulativeLimit,
            configured: true
        });
        emit LimitsUpdated(msg.sender, perTxLimit, dailyLimit, cumulativeLimit);
    }

    /// @notice Get the spending config for an account.
    function getConfig(address account) external view returns (SpendingConfig memory) {
        return configs[account];
    }

    /// @notice Get the spend state for a session key under an account.
    function getSpendState(address account, address sessionKey) external view returns (SpendState memory) {
        return spendStates[account][sessionKey];
    }

    // ---- ERC-165 ----

    function supportsInterface(bytes4 interfaceId) external pure returns (bool) {
        return interfaceId == 0x01ffc9a7; // ERC-165
    }

    // ---- Internal ----

    /// @dev Reset daily spend if the current day window has elapsed.
    function _resetDailyIfNeeded(SpendState storage state) internal {
        uint256 currentDay = block.timestamp / DAY;
        uint256 lastDay = state.dailyResetTimestamp / DAY;
        if (currentDay > lastDay) {
            state.dailySpent = 0;
            state.dailyResetTimestamp = block.timestamp;
        }
    }
}
