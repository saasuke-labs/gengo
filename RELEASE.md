# Release Process for Version 0.0.9

This document explains how to complete the release of version 0.0.9 after fixing the CI build error.

## Background

The initial v0.0.9 release failed in CI due to a build error. The issue was:
- GoReleaser was configured with `CGO_ENABLED=0`
- The dependency `github.com/chai2010/webp` (from `github.com/saasuke-labs/nagare@v0.0.8`) requires CGO for C code compilation
- This caused build failures on darwin_arm64 and other platforms with error: "undefined: webpGetInfo"

## Fixes Applied

The following changes were made to fix the build:
- ✅ Changed `CGO_ENABLED=0` to `CGO_ENABLED=1` in goreleaser.yaml
- ✅ Removed `-tags=netgo` flag (redundant with CGO enabled)
- ✅ Updated `.github/workflows/release.yaml` to use `goreleaser-cross:v1.24` container with cross-compilation toolchains

## Completing the Release

After merging the PR with the fixes to main, recreate the v0.0.9 tag:

### Steps to Recreate v0.0.9 Tag

```bash
# 1. Checkout main branch and pull latest changes
git checkout main
git pull origin main

# 2. Delete the old v0.0.9 tag locally and remotely
git tag -d v0.0.9
git push origin :refs/tags/v0.0.9

# 3. Create new v0.0.9 tag on the fixed commit
git tag v0.0.9

# 4. Push the new tag to trigger the release workflow
git push origin v0.0.9
```

### Alternative: Create Tag via GitHub Web Interface

1. Go to https://github.com/saasuke-labs/gengo/releases
2. Delete the existing v0.0.9 release and tag
3. Go to https://github.com/saasuke-labs/gengo/releases/new
4. Click "Choose a tag"
5. Type `v0.0.9` and select "Create new tag: v0.0.9 on publish"
6. Set the target to `main` (with the fixes merged)
7. Set the release title to "Release 0.0.9"
8. Click "Publish release"

## What Happens Next

Once the new tag is pushed:

1. GitHub Actions will detect the new tag matching pattern `v*.*.*`
2. The release workflow (`.github/workflows/release.yaml`) will trigger
3. The workflow will use the `goreleaser-cross:v1.24` container with all cross-compilation tools
4. GoReleaser will build binaries for multiple platforms with CGO enabled:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64, arm64)
5. Release artifacts will be uploaded to GitHub Releases
6. Checksums will be generated for verification

## Verification

After the workflow completes successfully, verify:
- All platform binaries are built without CGO errors
- Release assets are uploaded to https://github.com/saasuke-labs/gengo/releases/tag/v0.0.9
- No build errors related to `webpGetInfo` or other CGO issues
