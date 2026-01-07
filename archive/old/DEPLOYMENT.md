# Deployment Guide

This document provides a guide for deploying the Obsidian Automation Bot.

## 1. Docker (Recommended)

The most straightforward way to deploy the bot is using Docker and Docker Compose.

### Prerequisites
- Docker
- Docker Compose

### Steps
1.  **Create a `.env` file:**
    -   Copy the `.env.example` file to `.env`.
    -   Fill in the required environment variables, such as `TELEGRAM_BOT_TOKEN`, `GEMINI_API_KEYS`, etc.

2.  **Build and Run:**
    ```bash
    docker-compose up --build
    ```

    This command will build the Docker image and start the bot. To run in detached mode, use `docker-compose up -d --build`.

## 2. Kubernetes (Advanced)

For a more robust and scalable deployment, you can use Kubernetes.

### Prerequisites
- A running Kubernetes cluster
- `kubectl` configured to connect to your cluster

### Steps
1.  **Create a Secret:**
    ```bash
    kubectl create secret generic bot-secrets \
      --from-literal=TELEGRAM_BOT_TOKEN='your-token' \
      --from-literal=GEMINI_API_KEYS='your-keys' \
      --from-literal=...
    ```

2.  **Create a ConfigMap for `config.yml`:**
    ```bash
    kubectl create configmap bot-config --from-file=config.yml
    ```

3.  **Create a Deployment:**
    -   Create a `deployment.yaml` file (a sample is provided below).
    -   Apply it: `kubectl apply -f deployment.yaml`

### Sample `deployment.yaml`
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: obsidian-bot
spec:
  replicas: 1
  selector:
    matchLabels:
      app: obsidian-bot
  template:
    metadata:
      labels:
        app: obsidian-bot
    spec:
      containers:
      - name: bot
        image: your-docker-registry/obsidian-bot:latest
        envFrom:
        - secretRef:
            name: bot-secrets
        volumeMounts:
        - name: config
          mountPath: /app/config.yml
          subPath: config.yml
      volumes:
      - name: config
        configMap:
          name: bot-config
```

## 3. Deployment to a Cloud Provider (e.g., Google Cloud Run)

You can also deploy the bot as a serverless container on a cloud provider like Google Cloud Run.

### Prerequisites
- Google Cloud SDK (`gcloud`)
- A project on Google Cloud Platform with Cloud Run and Cloud Build APIs enabled

### Steps
1.  **Build and Push the Image:**
    ```bash
    gcloud builds submit --tag gcr.io/your-project-id/obsidian-bot
    ```

2.  **Deploy to Cloud Run:**
    ```bash
    gcloud run deploy obsidian-bot \
      --image gcr.io/your-project-id/obsidian-bot \
      --platform managed \
      --region your-region \
      --allow-unauthenticated \
      --set-env-vars "TELEGRAM_BOT_TOKEN=your-token,GEMINI_API_KEYS=your-keys"
    ```

---
**Note:** For all deployment methods, you will need to configure a webhook for the Telegram and WhatsApp APIs to point to the public URL of your deployed application.
