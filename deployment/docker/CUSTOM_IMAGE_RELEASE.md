# Custom Image Release

This repository is already prepared to publish the customized Semaphore server
image to GitHub Container Registry (GHCR).

## Recommended flow

1. Fork or mirror this repository into your GitHub account or organization.
2. Push your custom UI changes to that repository.
3. Publish a release tag like `overview-v6`.
4. Let GitHub Actions build and push the image to:

```text
ghcr.io/<github-owner>/semaphore-overview:overview-v6
```

The workflow file responsible for that is:

```text
.github/workflows/custom_ghcr_image.yml
```

## Triggering the image build

After committing your changes:

```bash
git tag overview-v6
git push origin overview-v6
```

You can also run the workflow manually through `Actions` on GitHub. In this
case, the workflow will publish branch and sha-based tags as well.

## Production server access

If the GitHub package is private, login once on the production server:

```bash
docker login ghcr.io
```

Use your GitHub username and a PAT with at least:

- `read:packages`

If the package is public, this login step is not required.

## Compose example

Update the Semaphore service image in production:

```yaml
services:
  semaphore:
    image: ghcr.io/<github-owner>/semaphore-overview:overview-v6
    restart: unless-stopped
```

After updating the compose, recreate the service:

```bash
docker compose up -d
```

If you manage the stack through 1Panel, just replace the `image:` field and
redeploy the compose.

## Rollback

Rollback is simply a matter of pointing the compose file to the previous stable
tag, for example:

```yaml
image: ghcr.io/<github-owner>/semaphore-overview:overview-v5
```
