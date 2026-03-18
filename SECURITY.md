# Security

## Reporting issues

Please report sensitive issues via GitHub **Security → Report a vulnerability** (private advisory) rather than a public issue.

## GitHub repository settings (once per repo)

1. **Dependabot** — **Settings → Code security and analysis** → enable **Dependabot alerts** and **Dependabot security updates**. The repo already includes [`.github/dependabot.yml`](.github/dependabot.yml) for version updates.

2. **Code scanning (CodeQL)** — Enable **Code scanning** so [`.github/workflows/codeql.yml`](.github/workflows/codeql.yml) results appear under **Security → Code scanning**. Public repositories include this free; private repos need **GitHub Advanced Security** for full integration.

3. **Secret scanning** — Enable if available on your plan; complements **gitleaks** in [`.github/workflows/security.yml`](.github/workflows/security.yml).

## CI checks

See the [README section on Security CI](README.md#security-ci-github-actions).
