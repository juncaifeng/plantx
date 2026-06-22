import { KitApiClient, createClient, type KitClientOptions } from '@plantx/kit-sdk-api';

export {
  PingRequest,
  PongResponse,
  EchoRequest,
  EchoResponse,
} from './generated/test-service/api/test';

import type {
  PingRequest,
  PongResponse,
  EchoRequest,
  EchoResponse,
} from './generated/test-service/api/test';

export class TestServiceClient {
  constructor(private client: KitApiClient) {}

  ping(_req: PingRequest = {}) {
    return this.client.get<PongResponse>('/test/v1/ping');
  }

  echo(req: EchoRequest) {
    return this.client.post<EchoResponse>('/test/v1/echo', req);
  }
}

export function createTestClient(options: KitClientOptions): TestServiceClient {
  return new TestServiceClient(createClient(options));
}
