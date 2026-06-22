import { KitApiClient, createClient, type KitClientOptions } from '@plantx/kit-sdk-api';

export interface Order {
  id: string;
  tenant_id: string;
  customer_name: string;
  status: string;
  created_at: number;
}

export interface CreateOrderRequest {
  customer_name: string;
}

export interface GetOrderRequest {
  id: string;
}

export interface ListOrdersRequest {
  status?: string;
}

export interface OrderList {
  orders: Order[];
}

export class OrderServiceClient {
  constructor(private client: KitApiClient) {}

  createOrder(req: CreateOrderRequest) {
    return this.client.post<Order>('/v1/orders', req);
  }

  getOrder(req: GetOrderRequest) {
    return this.client.get<Order>(`/v1/orders/${req.id}`);
  }

  listOrders(req: ListOrdersRequest = {}) {
    const qs = req.status ? `?status=${encodeURIComponent(req.status)}` : '';
    return this.client.get<OrderList>(`/v1/orders${qs}`);
  }
}

export function createOrderClient(options: KitClientOptions): OrderServiceClient {
  return new OrderServiceClient(createClient(options));
}
