export interface KitClientOptions {
  baseURL: string;
  getToken?: () => string | null;
  headers?: Record<string, string>;
  onUnauthorized?: () => void;
}

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  body?: unknown;
  headers?: Record<string, string>;
}

export class KitApiClient {
  constructor(private options: KitClientOptions) {}

  get baseURL(): string {
    return this.options.baseURL;
  }

  async request<T>(path: string, opts: RequestInit = {}): Promise<T> {
    const base = this.options.baseURL.replace(/\/$/, '');
    const normalizedPath = path.startsWith('/') ? path : `/${path}`;
    const url = `${base}${normalizedPath}`;
    const headers = new Headers(opts.headers);
    headers.set('Content-Type', 'application/json');
    const token = this.options.getToken?.();
    if (token) {
      headers.set('Authorization', `Bearer ${token}`);
    }
    const res = await fetch(url, { ...opts, headers });
    if (res.status === 401) {
      this.options.onUnauthorized?.();
    }
    if (!res.ok) {
      throw new Error(`HTTP ${res.status}: ${await res.text()}`);
    }
    return (await res.json()) as T;
  }

  get<T>(path: string) {
    return this.request<T>(path);
  }

  post<T>(path: string, body: unknown) {
    return this.request<T>(path, { method: 'POST', body: JSON.stringify(body) });
  }

  put<T>(path: string, body: unknown) {
    return this.request<T>(path, { method: 'PUT', body: JSON.stringify(body) });
  }

  patch<T>(path: string, body: unknown) {
    return this.request<T>(path, { method: 'PATCH', body: JSON.stringify(body) });
  }

  delete<T>(path: string) {
    return this.request<T>(path, { method: 'DELETE' });
  }
}

export { KitApiClient as KitAPIClient };
export { KitApiClient as ApiClient };

export function createClient(options: KitClientOptions): KitApiClient {
  return new KitApiClient(options);
}
