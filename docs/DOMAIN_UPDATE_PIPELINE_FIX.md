# Fix for Domain Update Pipeline Trigger Error

## Problem
The CI/CD pipeline for domain updates was failing with the following error:
```
pipeline trigger failed with status: 400, body: {"message":{"base":["deploy:domains-update job: chosen stage domain-update does not exist; available stages are .pre, tests, build, deploy, .post"]}}
```

## Root Cause
The `.gitlab-ci-domain-update.yml` file defined a custom stage `domain-update` that doesn't exist in the main `.gitlab-ci.yml` file. GitLab CI doesn't automatically merge stages from included files.

## Solution
Changed the job `deploy:domains-update` in `.gitlab-ci-domain-update.yml` to use the existing `deploy` stage instead of the non-existent `domain-update` stage.

### Changes Made:
1. **File: `.gitlab-ci-domain-update.yml`**
   - Removed the `stages:` section with `domain-update`
   - Changed `stage: domain-update` to `stage: deploy` for the `deploy:domains-update` job

## How It Works Now

### Pipeline Trigger Flow:
1. **Partner Create/Update/Delete** → Backend API detects change
2. **Backend API** → Calls `GitLabClient.TriggerDomainUpdateWithDetails()`
3. **GitLab API** → Triggers pipeline with variables:
   - `DOMAIN_UPDATE=true`
   - `DOMAIN_OPERATION=add|update|delete|refresh`
   - `OLD_DOMAIN=<old-domain>` (if applicable)
   - `NEW_DOMAIN=<new-domain>` (if applicable)
4. **GitLab CI** → Runs `deploy:domains-update` job in `deploy` stage
5. **Deployment Server** → Executes domain management scripts:
   - `manage-partner-domains.sh` - Updates nginx configs and SSL certificates
   - `update-monitoring-config.sh` - Updates monitoring
   - `health-check.sh` - Verifies everything works

## Testing

### 1. Test Pipeline Trigger Locally:
```bash
# Make sure environment variables are set
export GITLAB_API_URL=https://gitlab.com
export GITLAB_PROJECT_ID=<your-project-id>
export GITLAB_TRIGGER_TOKEN=<your-trigger-token>

# Test the trigger
./scripts/test-gitlab-pipeline-trigger.sh main add "" "newpartner.example.com"
```

### 2. Test via Backend API:
Create or update a partner through the admin API, and the pipeline should trigger automatically.

### 3. Validate CI Configuration:
```bash
./scripts/validate-gitlab-ci.sh
```

## Environment Variables Required

For the backend service to trigger pipelines:
- `GITLAB_API_URL` - GitLab instance URL (e.g., https://gitlab.com)
- `GITLAB_PROJECT_ID` - Project ID in GitLab
- `GITLAB_TRIGGER_TOKEN` - Pipeline trigger token from GitLab project settings

## Monitoring

Check pipeline status:
1. Go to GitLab project → CI/CD → Pipelines
2. Look for pipelines triggered by "Pipeline triggers API"
3. Check the `deploy:domains-update` job logs

## Troubleshooting

### If pipeline still fails:
1. Check that `GITLAB_TRIGGER_TOKEN` is set correctly in backend environment
2. Verify the trigger token is active in GitLab project settings
3. Check GitLab CI/CD settings allow pipeline triggers
4. Review the job logs for specific errors

### Common Issues:
- **Missing environment variables**: Ensure all required vars are set in GitLab CI/CD settings
- **SSH access issues**: Check `SSH_PRIVATE_KEY` is set correctly
- **Script permissions**: Ensure scripts are executable on the deployment server