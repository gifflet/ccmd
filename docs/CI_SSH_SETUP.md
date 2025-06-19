# CI/CD SSH Setup for Private Repositories

This document explains how to configure SSH access for the CI/CD pipeline to access private repositories during testing.

## GitHub Actions Setup

### 1. Generate SSH Key Pair

First, generate a new SSH key pair specifically for CI/CD:

```bash
ssh-keygen -t ed25519 -C "ccmd-ci@github.com" -f ccmd_ci_key
```

This will create two files:
- `ccmd_ci_key` (private key)
- `ccmd_ci_key.pub` (public key)

### 2. Add Deploy Key to Private Repository

1. Go to the private repository (e.g., `gifflet/hello-world`)
2. Navigate to Settings → Deploy keys
3. Click "Add deploy key"
4. Title: "CCMD CI/CD"
5. Key: Paste the contents of `ccmd_ci_key.pub`
6. Allow write access: No (read-only is sufficient)
7. Click "Add key"

### 3. Add Private Key to GitHub Secrets

1. Go to the `ccmd` repository
2. Navigate to Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Name: `SSH_PRIVATE_KEY`
5. Value: Paste the contents of `ccmd_ci_key` (including the BEGIN/END lines)
6. Click "Add secret"

### 4. Workflow Configuration

The workflow is already configured in `.github/workflows/ci.yml` to:
- Set up the SSH key from the secret
- Add GitHub to known hosts
- Configure the SSH agent
- Run tests that require private repository access

## Local Development Setup

For local development with private repositories:

1. Ensure your SSH key is added to your GitHub account
2. Test SSH access:
   ```bash
   ssh -T git@github.com
   ```
3. Run integration tests:
   ```bash
   go test ./tests/integration/...
   ```

## Troubleshooting

### SSH Key Not Working

1. Verify the key format (should include header/footer):
   ```
   -----BEGIN OPENSSH PRIVATE KEY-----
   ...
   -----END OPENSSH PRIVATE KEY-----
   ```

2. Check SSH agent:
   ```bash
   ssh-add -l
   ```

3. Test repository access:
   ```bash
   git clone git@github.com:gifflet/hello-world.git
   ```

### Permission Denied

- Ensure the deploy key is added to the correct repository
- Verify the SSH key in GitHub secrets matches the deploy key
- Check that the workflow is using the correct secret name

## Security Notes

### GitHub Actions Security

- **Secrets are automatically masked**: GitHub replaces any secret value with `***` in logs
- **This works even in public repositories**: Your SSH key will never be visible in logs
- **Fork PRs don't have access**: Pull requests from forks cannot access repository secrets
- **The workflow checks for this**: SSH setup only runs for pushes or PRs from the same repo

### Best Practices

- Use deploy keys instead of personal SSH keys for CI/CD
- Limit deploy key access to read-only when possible
- Rotate keys periodically
- Never commit private keys to the repository
- Use ED25519 keys for better security: `ssh-keygen -t ed25519`
- Name your keys descriptively (e.g., `ccmd-ci-key`)

### Additional Security Measures in Our Setup

1. **Output suppression**: The workflow redirects sensitive command outputs to `/dev/null`
2. **Conditional execution**: SSH setup only runs when the secret exists
3. **Minimal logging**: Only success/failure messages are shown, not key contents