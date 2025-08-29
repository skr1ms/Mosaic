#!/bin/bash

# Test script to verify GitLab pipeline triggering works correctly
# This simulates what the backend API does when triggering domain updates

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Testing GitLab Pipeline Trigger for Domain Updates${NC}"
echo "=================================================="

# Check required environment variables
if [ -z "$GITLAB_API_URL" ]; then
    echo -e "${RED}Error: GITLAB_API_URL is not set${NC}"
    exit 1
fi

if [ -z "$GITLAB_PROJECT_ID" ]; then
    echo -e "${RED}Error: GITLAB_PROJECT_ID is not set${NC}"
    exit 1
fi

if [ -z "$GITLAB_TRIGGER_TOKEN" ]; then
    echo -e "${RED}Error: GITLAB_TRIGGER_TOKEN is not set${NC}"
    exit 1
fi

# Default values
BRANCH="${1:-main}"
OPERATION="${2:-refresh}"
OLD_DOMAIN="${3:-}"
NEW_DOMAIN="${4:-test.example.com}"

echo "Configuration:"
echo "  API URL: $GITLAB_API_URL"
echo "  Project ID: $GITLAB_PROJECT_ID"
echo "  Branch: $BRANCH"
echo "  Operation: $OPERATION"
echo "  Old Domain: $OLD_DOMAIN"
echo "  New Domain: $NEW_DOMAIN"
echo ""

# Build the trigger URL
TRIGGER_URL="${GITLAB_API_URL}/api/v4/projects/${GITLAB_PROJECT_ID}/trigger/pipeline"

# Build form data
FORM_DATA="token=${GITLAB_TRIGGER_TOKEN}"
FORM_DATA="${FORM_DATA}&ref=${BRANCH}"
FORM_DATA="${FORM_DATA}&variables[DOMAIN_UPDATE]=true"
FORM_DATA="${FORM_DATA}&variables[DOMAIN_OPERATION]=${OPERATION}"

if [ -n "$OLD_DOMAIN" ]; then
    FORM_DATA="${FORM_DATA}&variables[OLD_DOMAIN]=${OLD_DOMAIN}"
fi

if [ -n "$NEW_DOMAIN" ]; then
    FORM_DATA="${FORM_DATA}&variables[NEW_DOMAIN]=${NEW_DOMAIN}"
fi

echo "Triggering pipeline..."
echo "URL: $TRIGGER_URL"
echo ""

# Make the API call
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "$FORM_DATA" \
    "$TRIGGER_URL")

# Extract HTTP status code
HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
BODY=$(echo "$RESPONSE" | head -n -1)

echo "Response:"
echo "  HTTP Status: $HTTP_CODE"
echo "  Body: $BODY"
echo ""

# Check if successful
if [ "$HTTP_CODE" -eq 201 ]; then
    echo -e "${GREEN}✅ Pipeline triggered successfully!${NC}"
    
    # Parse pipeline ID and URL from response
    PIPELINE_ID=$(echo "$BODY" | grep -o '"id":[0-9]*' | cut -d':' -f2)
    PIPELINE_URL=$(echo "$BODY" | grep -o '"web_url":"[^"]*' | cut -d'"' -f4)
    
    if [ -n "$PIPELINE_ID" ]; then
        echo "  Pipeline ID: $PIPELINE_ID"
    fi
    
    if [ -n "$PIPELINE_URL" ]; then
        echo "  Pipeline URL: $PIPELINE_URL"
    fi
else
    echo -e "${RED}❌ Failed to trigger pipeline!${NC}"
    echo "  Error details: $BODY"
    exit 1
fi

echo ""
echo "Test completed successfully!"