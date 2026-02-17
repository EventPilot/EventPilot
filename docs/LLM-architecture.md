# LLM Design Document

This document describes the LLM implementation for EventPilot using ClaudeAI. Users will interact with the agent through a chat window and can provide information to be turned into a social media post.

## Model Comparison CHart

| Model Name        | Output Price | Use case                                |
| ----------------- | ------------ | --------------------------------------- |
| Claude Sonnet 4.6 | $15/Mtok     | Balance of intelligence, cost and speed |
| Claude Opus 4.6   | $25/Mtok     | Most intelligence model for coding      |
| Claude Haiku 4.5  | $5/Mtok      | Fastest and most cost effective         |

## Why we went with Haiku 4.5
The Haiku 4.5 model has decent performance for our use case while using minimum tokens to save on cost.

## LLM Workflow
- Gather given information from database to determine what information is missing
- Use Anthropic API interfaced onto a chat window to communicate with user
- Prompt the user to provide missing event fields (event highlights, turnout, etc.)
- Collect data provided from user to fill in the missing information fields
- Prompt user to ensure all information is correct
- Call Bluesky API to create a Bluesky post under the user's handle