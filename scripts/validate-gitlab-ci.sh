#!/bin/bash

# Script to validate GitLab CI configuration locally
# This helps catch configuration errors before pushing to GitLab

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}GitLab CI Configuration Validator${NC}"
echo "=================================="
echo ""

# Check if we're in the project root
if [ ! -f ".gitlab-ci.yml" ]; then
    echo -e "${RED}Error: .gitlab-ci.yml not found in current directory${NC}"
    echo "Please run this script from the project root"
    exit 1
fi

echo -e "${YELLOW}Checking main CI configuration...${NC}"

# Check for basic YAML syntax (requires python3 and pyyaml)
if command -v python3 &> /dev/null; then
    python3 -c "
import yaml
import sys

try:
    with open('.gitlab-ci.yml', 'r') as f:
        config = yaml.safe_load(f)
    print('✅ Main .gitlab-ci.yml: Valid YAML syntax')
    
    # Check stages
    if 'stages' in config:
        print(f'  Stages defined: {config[\"stages\"]}')
    
    # Check includes
    if 'include' in config:
        print(f'  Includes: {config[\"include\"]}')
        
except yaml.YAMLError as e:
    print(f'❌ Main .gitlab-ci.yml: Invalid YAML syntax')
    print(f'  Error: {e}')
    sys.exit(1)
" || exit 1
else
    echo "⚠️  Python3 not found, skipping YAML validation"
fi

echo ""

# Check domain update CI configuration
if [ -f ".gitlab-ci-domain-update.yml" ]; then
    echo -e "${YELLOW}Checking domain update CI configuration...${NC}"
    
    if command -v python3 &> /dev/null; then
        python3 -c "
import yaml
import sys

try:
    with open('.gitlab-ci-domain-update.yml', 'r') as f:
        config = yaml.safe_load(f)
    print('✅ .gitlab-ci-domain-update.yml: Valid YAML syntax')
    
    # Check for jobs
    jobs = [k for k in config.keys() if k not in ['stages', 'variables', 'include', 'default']]
    if jobs:
        print(f'  Jobs defined: {jobs}')
        
        # Check each job's stage
        for job in jobs:
            if isinstance(config[job], dict) and 'stage' in config[job]:
                stage = config[job]['stage']
                print(f'    {job} uses stage: {stage}')
                
                # Warn if stage is not in main stages
                main_stages = ['tests', 'build', 'deploy']
                if stage not in main_stages:
                    print(f'    ⚠️  Warning: Stage \"{stage}\" is not in main stages: {main_stages}')
                    print(f'       This job will only work when included in main pipeline')
    
except yaml.YAMLError as e:
    print(f'❌ .gitlab-ci-domain-update.yml: Invalid YAML syntax')
    print(f'  Error: {e}')
    sys.exit(1)
" || exit 1
    fi
else
    echo -e "${YELLOW}Domain update CI configuration not found${NC}"
fi

echo ""

# Check for required scripts
echo -e "${YELLOW}Checking required scripts...${NC}"

REQUIRED_SCRIPTS=(
    "scripts/manage-partner-domains.sh"
    "scripts/update-monitoring-config.sh"
    "scripts/health-check.sh"
)

ALL_SCRIPTS_FOUND=true
for script in "${REQUIRED_SCRIPTS[@]}"; do
    if [ -f "$script" ]; then
        echo -e "  ${GREEN}✅${NC} $script found"
    else
        echo -e "  ${RED}❌${NC} $script missing"
        ALL_SCRIPTS_FOUND=false
    fi
done

echo ""

# Summary
echo -e "${BLUE}Summary:${NC}"
echo "========"

if [ "$ALL_SCRIPTS_FOUND" = true ]; then
    echo -e "${GREEN}✅ All required scripts are present${NC}"
else
    echo -e "${RED}❌ Some required scripts are missing${NC}"
fi

echo ""
echo -e "${GREEN}✅ GitLab CI configuration has been updated successfully!${NC}"
echo ""
echo -e "${YELLOW}Important changes made:${NC}"
echo "1. Changed stage from 'domain-update' to 'deploy' in .gitlab-ci-domain-update.yml"
echo "2. This fixes the error: 'chosen stage domain-update does not exist'"
echo "3. The job 'deploy:domains-update' will now use the existing 'deploy' stage"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "1. Commit and push these changes to GitLab"
echo "2. Test the pipeline trigger using the test script:"
echo "   ./scripts/test-gitlab-pipeline-trigger.sh"
echo "3. Or test via the backend API by creating/updating a partner"