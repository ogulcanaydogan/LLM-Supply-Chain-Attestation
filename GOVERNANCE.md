# Governance

This document describes how the LLM Supply-Chain Attestation project is maintained, how decisions are made, and how contributors can participate.

## Project Status

This project is in active development. The API and attestation schema are stabilising toward a 1.0 release. Breaking changes are documented in [CHANGELOG.md](CHANGELOG.md) and follow semantic versioning.

## Core Maintainers

| Name | GitHub | Role |
|------|--------|------|
| Ogulcan Aydogan | [@ogulcanaydogan](https://github.com/ogulcanaydogan) | Project lead, primary maintainer |

Core maintainers have commit access, release authority, and final decision-making power on architectural direction.

## Becoming a Maintainer

Contributors who demonstrate sustained, high-quality contributions may be invited to become maintainers. Criteria include:

- Multiple accepted pull requests across different areas of the codebase.
- Participation in issue triage and code review.
- Understanding of the project's security model and attestation architecture.
- Alignment with the project's goals and code of conduct.

## Decision-Making Process

### Day-to-Day Decisions

Routine decisions (bug fixes, minor enhancements, documentation improvements) are made by any core maintainer through the standard pull request process.

### Major Decisions

Architectural changes, new attestation types, breaking schema changes, and dependency additions require:

1. A GitHub issue or discussion describing the proposal.
2. Review period of at least 7 days for community feedback.
3. Approval from at least one core maintainer.
4. Documentation updates in README, CHANGELOG, and relevant guides.

Examples of major decisions:
- Adding a new attestation type (e.g., `deployment_attestation`).
- Changing the DSSE envelope format or bundle schema.
- Adding a new signing provider (KMS, HSM).
- Modifying the verification engine's trust model.

## Release Cadence

- **Patch releases** (0.x.Y): As needed for bug fixes and security patches.
- **Minor releases** (0.X.0): Monthly or when significant features are complete.
- **Major release** (1.0.0): When the attestation schema and CLI interface are stable.

Releases are automated via GoReleaser and published through GitHub Actions. See [CONTRIBUTING.md](CONTRIBUTING.md) for the release process.

## Code of Conduct

All participants are expected to:

- Be respectful and constructive in discussions and code reviews.
- Focus on technical merit when evaluating contributions.
- Welcome newcomers and help them understand the project.
- Report unacceptable behaviour to the core maintainers.

## Contribution Model

This project follows a fork-and-pull-request model:

1. Fork the repository.
2. Create a feature branch from `main`.
3. Submit a pull request with a clear description.
4. Address review feedback.
5. A core maintainer merges after CI passes and review is complete.

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed development setup and guidelines.

## Communication Channels

- **GitHub Issues**: Bug reports, feature requests, and general questions.
- **GitHub Discussions**: Architecture proposals and design conversations.
- **Pull Requests**: Code contributions and review.
- **Security Advisories**: Vulnerability reports (see [SECURITY.md](SECURITY.md)).
