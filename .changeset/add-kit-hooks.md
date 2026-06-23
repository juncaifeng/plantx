---
"@plantx/kit-sdk-kit": minor
---

Add data-fetching hooks for platform registry resources:

- `useApplications` – list and filter active applications
- `useMenus` – list menus globally or for a specific application
- `useMicroApps` – list micro-apps globally or for a specific application
- `useRegistryClient` – reusable registry service client helper

The previous `useMicroApps(manifests)` stub has been replaced with a
registry-backed implementation that loads micro-app manifests from
`registry-service`.
