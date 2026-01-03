# Plan: AI Provider Expansion and Dynamic Selection

## Phase 1: AI Provider Integration
- [x] Task: Implement Hugging Face AIProvider 984a7ad
  - [x] Sub-task: Define Hugging Face API client (internal/ai/huggingface_provider.go) 984a7ad
  - [x] Sub-task: Implement GetCompletion and GetCompletionStream methods for Hugging Face 984a7ad
  - [x] Sub-task: Add Hugging Face provider configuration to config.yml (model, API key) 984a7ad
  - [x] Sub-task: Integrate Hugging Face provider into AIService (internal/ai/ai_service.go) 984a7ad
- [x] Task: Implement OpenRouter AIProvider a5397d0
  - [x] Sub-task: Define OpenRouter API client (internal/ai/openrouter_provider.go) a5397d0
  - [x] Sub-task: Implement GetCompletion and GetCompletionStream methods for OpenRouter a5397d0
  - [x] Sub-task: Add OpenRouter provider configuration to config.yml (model, API key) a5397d0
  - [x] Sub-task: Integrate OpenRouter provider into AIService (internal/ai/ai_service.go) a5397d0
- [x] Task: Conductor - User Manual Verification 'Phase 1: AI Provider Integration' (Protocol in workflow.md) 713d013

## Phase 2: Dynamic Provider Selection Enhancement
- [x] Task: Implement Provider Health Check Mechanism 713d013
  - [x] Sub-task: Define a method to check the health/status of each AIProvider (e.g., in internal/ai/provider.go or a new health package) 713d013
  - [x] Sub-task: Integrate health check into AIService to retrieve status of all providers 713d013
- [x] Task: Enhance Telegram /setprovider Command 713d013
  - [x] Sub-task: Modify command handler (internal/bot/main.go) to retrieve active/healthy providers from AIService 713d013
  - [x] Sub-task: Generate dynamic inline keyboard sub-menu for provider selection based on health status 713d013
  - [x] Sub-task: Implement callback query handling for sub-menu selections to set the active provider 713d013
- [x] Task: Conductor - User Manual Verification 'Phase 2: Dynamic Provider Selection Enhancement' (Protocol in workflow.md) 713d013
