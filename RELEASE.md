# Release Process for Version 0.0.8

This document explains how to complete the release of version 0.0.8.

## Current Status

- ✅ CHANGELOG.md has been created documenting version 0.0.8
- ✅ Local annotated tag `v0.0.8` has been created
- ⏳ Tag needs to be pushed by a maintainer with appropriate permissions

## Completing the Release

The tag `v0.0.8` has been created locally but cannot be pushed due to repository protection rules that restrict tag creation. To complete the release, a maintainer with appropriate permissions should:

### Option 1: Create Tag via GitHub Web Interface

1. Go to https://github.com/saasuke-labs/gengo/releases/new
2. Click "Choose a tag"
3. Type `v0.0.8` and select "Create new tag: v0.0.8 on publish"
4. Set the target to the branch `copilot/release-gengo-version-008` (or `main` after merging)
5. Set the release title to "Release 0.0.8"
6. Copy the content from CHANGELOG.md for version 0.0.8 into the release description
7. Click "Publish release"

### Option 2: Create Tag via Git (requires permissions)

```bash
# Ensure you're on the correct branch
git checkout copilot/release-gengo-version-008  # or main after merging

# Pull the latest changes
git pull

# Create and push the annotated tag
git tag -a v0.0.8 -m "Release version 0.0.8"
git push origin v0.0.8
```

### Option 3: Create Tag via GitHub CLI (requires permissions)

```bash
gh release create v0.0.8 --title "Release 0.0.8" --notes-file CHANGELOG.md --target copilot/release-gengo-version-008
```

## What Happens Next

Once the tag is pushed:

1. GitHub Actions will detect the new tag matching pattern `v*.*.*`
2. The release workflow (`.github/workflows/release.yaml`) will trigger
3. GoReleaser will build binaries for multiple platforms (Linux, macOS, Windows)
4. Release artifacts will be uploaded to GitHub Releases
5. Checksums will be generated for verification

## Repository Protection Rules

The repository has protection rules that prevent direct tag creation via git push. This is a security measure to ensure releases are controlled and authorized. Tags that trigger releases should only be created by maintainers with appropriate permissions.
