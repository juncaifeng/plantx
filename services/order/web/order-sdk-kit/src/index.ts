import { useCallback, useEffect, useState } from 'react';
import {
  createOrderClient,
  OrderServiceClient,
  type Order,
  type CreateOrderRequest,
} from '@plantx/order-sdk-api';
import { useKitContext } from '@plantx/kit-sdk-kit';

export interface UseOrdersOptions {
  status?: string;
}

export function useOrders() {
  const ctx = useKitContext();
  const [client] = useState<OrderServiceClient>(() =>
    createOrderClient({
      baseURL: ctx.apiClient?.baseURL ?? '/api/order',
    })
  );
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const list = useCallback(
    async (opts: UseOrdersOptions = {}) => {
      setLoading(true);
      try {
        const res = await client.listOrders(opts);
        setOrders(res.orders);
      } catch (e) {
        setError(e as Error);
      } finally {
        setLoading(false);
      }
    },
    [client]
  );

  const create = useCallback(
    async (req: CreateOrderRequest) => {
      const created = await client.createOrder(req);
      setOrders((prev: Order[]) => [created, ...prev]);
      return created;
    },
    [client]
  );

  useEffect(() => {
    list();
  }, [list]);

  return { orders, loading, error, list, create };
}
