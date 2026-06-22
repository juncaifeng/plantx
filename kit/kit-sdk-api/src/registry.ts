import type { KitApiClient } from './index.js';

export interface Route {
  path: string;
  method: string;
}

export interface MicroApp {
  name: string;
  route: string;
  bundleUrl: string;
  menuLabelKey: string;
  requirePermission: string;
  applicationId?: string;
  applicationKey?: string;
}

export interface Service {
  id: string;
  name: string;
  grpcHost: string;
  restPrefix: string;
  routes?: Route[];
  microApps?: MicroApp[];
  applicationId?: string;
  applicationKey?: string;
}

export interface RegisterServiceRequest {
  name: string;
  grpcHost: string;
  restPrefix: string;
  applicationId?: string;
  applicationKey?: string;
}

export interface DeregisterServiceRequest {
  id: string;
}

export interface GetServiceRequest {
  id: string;
}

export interface ListServicesRequest {}

export interface ServiceList {
  services: Service[];
}

export interface RegisterMicroAppRequest {
  serviceName: string;
  microApp: MicroApp;
  applicationId?: string;
  applicationKey?: string;
}

export interface UpdateMicroAppRequest {
  name: string;
  route?: string;
  bundleUrl?: string;
  menuLabelKey?: string;
  requirePermission?: string;
}

export interface DeleteMicroAppRequest {
  name: string;
}

export interface ListMicroAppsRequest {}

export interface MicroAppList {
  microApps: MicroApp[];
}

export interface Menu {
  id: string;
  labelKey: string;
  route?: string;
  icon?: string;
  parentId?: string;
  sortOrder: number;
  microAppName?: string;
  requirePermission?: string;
  applicationId?: string;
  applicationKey?: string;
}

export interface CreateMenuRequest {
  labelKey: string;
  route?: string;
  icon?: string;
  parentId?: string;
  sortOrder?: number;
  microAppName?: string;
  requirePermission?: string;
  applicationId?: string;
  applicationKey?: string;
}

export interface UpdateMenuRequest {
  id: string;
  labelKey?: string;
  route?: string;
  icon?: string;
  parentId?: string;
  sortOrder?: number;
  microAppName?: string;
  requirePermission?: string;
  applicationId?: string;
  applicationKey?: string;
}

export interface DeleteMenuRequest {
  id: string;
}

export interface ReorderMenusRequest {
  items: { id: string; sortOrder: number }[];
}

export interface ListMenusRequest {}

export interface MenuList {
  menus: Menu[];
}

export interface RoutePolicy {
  rateLimitRps?: number;
  authRequired?: boolean;
  canaryWeight?: number;
  canaryHost?: string;
}

export interface ServiceRoute {
  serviceId: string;
  name: string;
  restPrefix: string;
  upstreamHost: string;
  routes?: Route[];
  policy?: RoutePolicy;
}

export interface SyncRoutesRequest {}

export interface SyncRoutesResponse {
  routes: ServiceRoute[];
}

export interface GetRoutePolicyRequest {
  serviceId: string;
}

export interface SetRoutePolicyRequest {
  serviceId: string;
  policy: RoutePolicy;
}

export type ApplicationStatus =
  | 'APPLICATION_STATUS_ACTIVE'
  | 'APPLICATION_STATUS_OFFLINE'
  | 'APPLICATION_STATUS_UNSPECIFIED';

export interface Application {
  id: string;
  key: string;
  name: string;
  labelKey: string;
  icon?: string;
  description?: string;
  status: ApplicationStatus;
  sortOrder: number;
}

export interface RegisterApplicationRequest {
  key: string;
  name: string;
  labelKey: string;
  icon?: string;
  description?: string;
  status?: ApplicationStatus;
  sortOrder?: number;
}

export interface GetApplicationRequest {
  id: string;
}

export interface UpdateApplicationRequest {
  id: string;
  key?: string;
  name?: string;
  labelKey?: string;
  icon?: string;
  description?: string;
  status?: ApplicationStatus;
  sortOrder?: number;
}

export interface DeleteApplicationRequest {
  id: string;
}

export interface ListApplicationsRequest {}

export interface ApplicationList {
  applications: Application[];
}

export interface GetApplicationMenusRequest {
  applicationId: string;
}

export interface GetApplicationMicroAppsRequest {
  applicationId: string;
}

export class RegistryServiceClient {
  constructor(private readonly client: KitApiClient) {}

  registerService(body: RegisterServiceRequest): Promise<Service> {
    return this.client.post<Service>('registry/v1/services', body);
  }

  deregisterService(request: DeregisterServiceRequest): Promise<void> {
    return this.client.delete<void>(`registry/v1/services/${encodeURIComponent(request.id)}`);
  }

  getService(request: GetServiceRequest): Promise<Service> {
    return this.client.get<Service>(`registry/v1/services/${encodeURIComponent(request.id)}`);
  }

  listServices(_?: ListServicesRequest): Promise<ServiceList> {
    return this.client.get<ServiceList>('registry/v1/services');
  }

  registerMicroApp(body: RegisterMicroAppRequest): Promise<MicroApp> {
    return this.client.post<MicroApp>('registry/v1/micro-apps', body);
  }

  updateMicroApp(body: UpdateMicroAppRequest): Promise<MicroApp> {
    return this.client.put<MicroApp>(`registry/v1/micro-apps/${encodeURIComponent(body.name)}`, body);
  }

  deleteMicroApp(request: DeleteMicroAppRequest): Promise<void> {
    return this.client.delete<void>(`registry/v1/micro-apps/${encodeURIComponent(request.name)}`);
  }

  listMicroApps(_?: ListMicroAppsRequest): Promise<MicroAppList> {
    return this.client.get<MicroAppList>('registry/v1/micro-apps');
  }

  createMenu(body: CreateMenuRequest): Promise<Menu> {
    return this.client.post<Menu>('registry/v1/menus', body);
  }

  listMenus(_?: ListMenusRequest): Promise<MenuList> {
    return this.client.get<MenuList>('registry/v1/menus');
  }

  updateMenu(body: UpdateMenuRequest): Promise<Menu> {
    return this.client.put<Menu>(`registry/v1/menus/${encodeURIComponent(body.id)}`, body);
  }

  deleteMenu(request: DeleteMenuRequest): Promise<void> {
    return this.client.delete<void>(`registry/v1/menus/${encodeURIComponent(request.id)}`);
  }

  reorderMenus(body: ReorderMenusRequest): Promise<MenuList> {
    return this.client.post<MenuList>('registry/v1/menus/reorder', body);
  }

  syncRoutes(_?: SyncRoutesRequest): Promise<SyncRoutesResponse> {
    return this.client.get<SyncRoutesResponse>('registry/v1/sync-routes');
  }

  getRoutePolicy(request: GetRoutePolicyRequest): Promise<RoutePolicy> {
    return this.client.get<RoutePolicy>(`registry/v1/services/${encodeURIComponent(request.serviceId)}/route-policy`);
  }

  setRoutePolicy(request: SetRoutePolicyRequest): Promise<RoutePolicy> {
    return this.client.put<RoutePolicy>(`registry/v1/services/${encodeURIComponent(request.serviceId)}/route-policy`, request);
  }

  registerApplication(body: RegisterApplicationRequest): Promise<Application> {
    return this.client.post<Application>('registry/v1/applications', body);
  }

  listApplications(_?: ListApplicationsRequest): Promise<ApplicationList> {
    return this.client.get<ApplicationList>('registry/v1/applications');
  }

  getApplication(request: GetApplicationRequest): Promise<Application> {
    return this.client.get<Application>(`registry/v1/applications/${encodeURIComponent(request.id)}`);
  }

  updateApplication(body: UpdateApplicationRequest): Promise<Application> {
    return this.client.put<Application>(`registry/v1/applications/${encodeURIComponent(body.id)}`, body);
  }

  deleteApplication(request: DeleteApplicationRequest): Promise<void> {
    return this.client.delete<void>(`registry/v1/applications/${encodeURIComponent(request.id)}`);
  }

  getApplicationMenus(request: GetApplicationMenusRequest): Promise<MenuList> {
    return this.client.get<MenuList>(`registry/v1/applications/${encodeURIComponent(request.applicationId)}/menus`);
  }

  getApplicationMicroApps(request: GetApplicationMicroAppsRequest): Promise<MicroAppList> {
    return this.client.get<MicroAppList>(`registry/v1/applications/${encodeURIComponent(request.applicationId)}/micro-apps`);
  }
}
