#!/bin/bash

# Script for preloading Docker images to GitLab Runner
# Run this script on the server where GitLab Runner is installed

echo "🚀 Preparing Docker images for CI/CD..."

# Base images for build processes
echo "📥 Downloading base images..."
docker pull golang:1.24-alpine
docker pull node:20-alpine
docker pull nginx:alpine
docker pull docker:latest
docker pull ubuntu:22.04

# Images for services (database, cache)
echo "📥 Downloading images for services..."
docker pull postgres:17-alpine
docker pull redis:8-alpine

echo "✅ All images downloaded! List:"
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"

echo ""
echo "🔧 Setup completed! Now GitLab CI will not access Docker Hub."
