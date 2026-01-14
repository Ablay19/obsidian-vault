# Architecture Documentation

This directory contains architecture documentation for the Obsidian Vault project.

## Contents

- [Architecture Overview](./ARCHITECTURE.md) - Main architecture documentation

## Architecture Separation

The project follows a microservices architecture with:

- **Go Applications** (`apps/`) - Backend services
- **JavaScript Workers** (`workers/`) - Cloudflare Workers for edge processing
- **Shared Packages** (`packages/`) - Common type definitions and utilities
- **Deployment** (`deploy/`) - Docker, Kubernetes, and Terraform configurations

## Quick Links

- [API Gateway Documentation](./apps/api-gateway/README.md)
- [Worker Documentation](./workers/ai-worker/README.md)
- [Deployment Guide](../DEPLOYMENT.md)
