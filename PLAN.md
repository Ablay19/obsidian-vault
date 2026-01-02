# ONNX Runtime Integration Plan

This document outlines the steps to integrate ONNX Runtime with the Go application, enabling local AI inference using ONNX models.

## Phase 1: Environment Setup

1.  **Install ONNX Runtime C Library:**
    *   Download the ONNX Runtime C library for Linux.
    *   Extract and move it to `/opt/onnxruntime`.
    *   Set `LD_LIBRARY_PATH`, `CGO_CFLAGS`, and `CGO_LDFLAGS` environment variables to point to the ONNX Runtime installation.

2.  **Add Go Bindings:**
    *   Add `github.com/yalue/onnxruntime_go` to the Go project dependencies using `go get`.
    *   Run `go mod tidy` to clean up dependencies.

## Phase 2: Code Implementation

1.  **Create ONNX Provider (`internal/ai/onnx_provider.go`):**
    *   Implement the `AIProvider` interface for ONNX.
    *   Include logic for initializing ONNX sessions, loading models, and running inference.
    *   Add placeholder functions for `prepareInput` and `processOutput` which will be model-specific.

2.  **Create Simple Tokenizer (`internal/ai/onnx_tokenizer.go`):**
    *   Implement a basic tokenizer to convert text prompts into a format suitable for ONNX models. This will need to be adapted for specific models.

3.  **Update AI Service (`internal/ai/ai_service.go`):**
    *   Modify the `NewAIService` function to include initialization of the `ONNXProvider`.
    *   Register the ONNX provider in the `providers` map.

## Phase 3: Model and Configuration

1.  **Download a Test ONNX Model:**
    *   Acquire a suitable ONNX model (e.g., a text classification model like DistilBERT) and place it in the `models` directory within the project.

2.  **Update Configuration (`config.yaml` or `.env`):**
    *   Add configuration entries for the ONNX provider, including the `model_path` and an `enabled` flag.

## Phase 4: Build and Dockerization

1.  **Build with ONNX Support:**
    *   Ensure all necessary environment variables for CGO are set.
    *   Build the Go application.

2.  **Create Dockerfile for ONNX (`Dockerfile.onnx`):
    *   Develop a multi-stage Dockerfile that:
        *   Installs ONNX Runtime C library during the build stage.
        *   Sets up the environment for CGO compilation.
        *   Builds the Go application.
        *   In the runtime stage, copies the compiled binary, ONNX Runtime libraries, and the ONNX model.

## Phase 5: Testing

1.  **Run Local Build:**
    *   Execute the locally built binary to verify ONNX integration.

2.  **Run Dockerized Build:**
    *   Build the `Dockerfile.onnx` image.
    *   Run the Docker container, mapping necessary environment variables and volumes.
    *   Test the bot's functionality with the ONNX provider.
