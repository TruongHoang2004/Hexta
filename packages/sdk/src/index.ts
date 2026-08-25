import axios, { AxiosInstance, AxiosRequestConfig } from "axios";
import Cookies from "js-cookie";

export interface Tenant {
  id: string;
  name: string;
  slug: string;
  status: string;
  plan: string;
  owner_id: string;
}

export interface User {
  id: string;
  tenant_id: string;
  email: string;
  full_name: string;
  status: string;
  role_id: string;
}

export interface StorageAdapter {
  get: (key: string) => string | null | Promise<string | null>;
  set: (key: string, value: string) => void | Promise<void>;
  remove: (key: string) => void | Promise<void>;
}

export class DefaultBrowserStorage implements StorageAdapter {
  get(key: string): string | null {
    if (typeof window !== "undefined") {
      return Cookies.get(key) || null;
    }
    return null;
  }
  set(key: string, value: string): void {
    if (typeof window !== "undefined") {
      Cookies.set(key, value, { expires: 7, path: "/" });
    }
  }
  remove(key: string): void {
    if (typeof window !== "undefined") {
      Cookies.remove(key, { path: "/" });
    }
  }
}

export class IdentityClient {
  private client: AxiosInstance;
  private storage: StorageAdapter;
  private authKey = "auth_token";

  constructor(client: AxiosInstance, storage: StorageAdapter) {
    this.client = client;
    this.storage = storage;
  }

  // Auth
  async login(email: string, password?: string): Promise<{ token: string }> {
    const response = await this.client.post<{ data: { access_token: string } }>("/api/v1/auth/login", { email, password });
    
    // We expect the standard response format from the API (Response[T])
    const token = response.data.data.access_token;
    if (token) {
      await this.storage.set(this.authKey, token);
    }
    
    return { token };
  }

  async register(email: string, password?: string): Promise<{ token: string }> {
    const response = await this.client.post<{ data: { access_token: string } }>("/api/v1/auth/register", { email, password });
    
    const token = response.data.data.access_token;
    if (token) {
      await this.storage.set(this.authKey, token);
    }
    
    return { token };
  }

  async logout(): Promise<void> {
    try {
      await this.client.post("/api/v1/auth/logout");
    } catch (e) {
      // Ignore errors on logout (e.g. already logged out)
    } finally {
      await this.storage.remove(this.authKey);
    }
  }

  // Tenant
  async getTenant(id: string): Promise<Tenant> {
    const response = await this.client.get<{ data: Tenant }>(`/api/v1/tenants/${id}`);
    return response.data.data;
  }

  // Users
  async getUsers(tenantId: string): Promise<User[]> {
    const response = await this.client.get<{ data: User[] }>(`/api/v1/tenants/${tenantId}/users`);
    return response.data.data;
  }
}

export interface SDKConfig {
  apiUrl: string;
  storage?: StorageAdapter;
}

export class EcommerceHubSDK {
  public identity: IdentityClient;
  private client: AxiosInstance;
  private storage: StorageAdapter;

  constructor(config: SDKConfig) {
    this.storage = config.storage || new DefaultBrowserStorage();
    
    this.client = axios.create({
      baseURL: config.apiUrl,
      headers: {
        "Content-Type": "application/json",
      },
    });

    // Request interceptor to inject token
    this.client.interceptors.request.use(async (reqConfig) => {
      const token = await this.storage.get("auth_token");
      if (token) {
        reqConfig.headers = reqConfig.headers || {};
        reqConfig.headers.Authorization = `Bearer ${token}`;
      }
      return reqConfig;
    });

    // Response interceptor to handle standard API errors
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response && error.response.data) {
          const apiError = error.response.data;
          throw new Error(apiError.message || apiError.detail || "API Error");
        }
        throw error;
      }
    );

    this.identity = new IdentityClient(this.client, this.storage);
  }
}
