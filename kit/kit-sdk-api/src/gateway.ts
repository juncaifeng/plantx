import { KitApiClient } from './index.js';

export interface Route {
  path: string;
  method: string;
}

export interface MicroApp {
  name: string;
  route: string;
  bundle_url: string;
  menu_label_key: string;
  require_permission: string;
}

export interface Service {
  id: string;
  name: string;
  grpc_host: string;
  rest_prefix: string;
  routes?: Route[];
  micro_app?: MicroApp;
}

export interface RegisterServiceRequest {
  name: string;
  grpc_host: string;
  rest_prefix: string;
}

export interface ListServicesRequest {}

export interface ServiceList {
  services: Service[];
}

export interface ListRoutesRequest {
  id: string;
}

export interface RouteList {
  routes: Route[];
}

export interface RegisterMicroAppRequest {
  service_name: string;
  micro_app: MicroApp;
}

export interface ListMicroAppsRequest {}

export interface MicroAppList {
  micro_apps: MicroApp[];
}

export class GatewayServiceClient {
  constructor(private readonly client: KitApiClient) {}

  registerService(body: RegisterServiceRequest): Promise<Service> {
    return this.client.post<Service>('/gateway/v1/services', body);
  }

  listServices(_?: ListServicesRequest): Promise<ServiceList> {
    return this.client.get<ServiceList>('/gateway/v1/services');
  }

  listRoutes(request: ListRoutesRequest): Promise<RouteList> {
    return this.client.get<RouteList>(`/gateway/v1/services/${encodeURIComponent(request.id)}/routes`);
  }

  registerMicroApp(body: RegisterMicroAppRequest): Promise<MicroApp> {
    return this.client.post<MicroApp>('/gateway/v1/micro-apps', body);
  }

  listMicroApps(_?: ListMicroAppsRequest): Promise<MicroAppList> {
    return this.client.get<MicroAppList>('/gateway/v1/micro-apps');
  }
}
