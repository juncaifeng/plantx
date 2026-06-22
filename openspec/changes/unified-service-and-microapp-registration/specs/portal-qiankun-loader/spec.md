## ADDED Requirements

### Requirement: Portal loads child apps via qiankun
`apps/portal/src/MicroAppPage.tsx` SHALL use `qiankun`'s `loadMicroApp` to mount and unmount child apps, replacing the current custom `<script>` injection.

#### Scenario: User navigates to a micro-app route
- **WHEN** the user visits `/test`
- **THEN** portal calls `loadMicroApp({ name, entry, container, props })` and the child app renders inside the container

### Requirement: Portal unmounts child apps on route leave
When the user navigates away from a micro-app route, `MicroAppPage` SHALL unmount the qiankun micro-app instance.

#### Scenario: User leaves the micro-app route
- **WHEN** the React component unmounts or route changes
- **THEN** portal calls `microApp.unmount()` and cleans up the container

### Requirement: Portal injects kit context through qiankun props
Portal SHALL pass `user`, `tenant`, `permissions`, `apiClient`, and `locale` to the child app via qiankun props.

#### Scenario: test-ui receives context
- **WHEN** `test-ui` is mounted
- **THEN** its `mount(props)` receives `props.user`, `props.tenant`, `props.permissions`, `props.apiClient`, and `props.locale`
