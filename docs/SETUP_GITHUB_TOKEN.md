# Setup GitHub Token for Projects API

The `gh auth refresh` command doesn't always work reliably. Here's the manual approach:

## Step 1: Create Personal Access Token

1. Go to: https://github.com/settings/tokens/new
2. Token name: "Ubik Projects Automation"
3. Expiration: 90 days (or No expiration)
4. Select these scopes:
   - ✅ `repo` (Full control of private repositories)
   - ✅ `workflow` (Update GitHub Action workflows)
   - ✅ `project` (Full control of projects) ← **IMPORTANT!**
   - ✅ `read:org` (Read org and team membership)

5. Click "Generate token"
6. **Copy the token** (starts with `ghp_...`)

## Step 2: Configure gh CLI with new token

```bash
# Set the token as environment variable
export GH_TOKEN="ghp_YOUR_TOKEN_HERE"

# Test it works
gh auth status

# Or login with the token
echo "ghp_YOUR_TOKEN_HERE" | gh auth login --with-token
```

## Step 3: Verify project access

```bash
# This should now work
gh project list

# And this should work
gh api graphql -f query='query { viewer { login } }'
```

## Step 4: Run setup script

```bash
./scripts/setup-project-automation.sh
```

---

## Alternative: Use REST API endpoint

If you don't want to create a new token, you can get project info via browser:

1. Go to: https://github.com/users/sergei-rastrigin/projects
2. Click on each project
3. Note the project number in URL: `/users/sergei-rastrigin/projects/1`
4. Manually create config file (see below)

### Manual config file:

Create `.github/project-config.json`:

```json
{
  "projects": {
    "engineering": {
      "id": "PROJECT_ID_HERE",
      "number": 1,
      "title": "Ubik Engineering Roadmap"
    },
    "marketing": {
      "id": "PROJECT_ID_HERE",
      "number": 2,
      "title": "Ubik Business & Marketing"
    }
  }
}
```

To get project IDs, you'll need the GraphQL API (which requires the token).

---

## Recommended: Use Personal Access Token

The manual token approach is most reliable for automation.
