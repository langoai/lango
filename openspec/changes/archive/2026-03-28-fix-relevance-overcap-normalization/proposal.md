# Proposal: Fix relevance score over-cap normalization

## Problem

The two-step BoostRelevanceScore cap step used `RelevanceScoreLT(maxScore)` which excluded pre-existing over-cap values (e.g., score=5.10 when maxScore=5.0) from normalization. These values were created by the prior single-step boost bug and remain in the database after upgrade.

## Fix

Remove the `RelevanceScoreLT(maxScore)` upper bound from the cap step. The condition becomes `RelevanceScoreGT(maxScore-delta)` only, which catches both overshoot prevention (scores near the cap) AND pre-existing over-cap normalization (scores already above maxScore).

Added test case: `pre-existing over-cap normalized: 5.10 + 0.05 -> 5.00`.
