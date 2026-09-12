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

export interface TokenPair {
  access_token: string;
  refresh_token?: string;
}

export interface StorageOptions {
  expires?: number | Date;
  path?: string;
  domain?: string;
  secure?: boolean;
  sameSite?: "strict" | "Strict" | "lax" | "Lax" | "none" | "None";
}

export interface StorageAdapter {
  get: (key: string) => string | null | Promise<string | null>;
  set: (key: string, value: string, options?: StorageOptions) => void | Promise<void>;
  remove: (key: string, options?: StorageOptions) => void | Promise<void>;
}

export class DefaultBrowserStorage implements StorageAdapter {
  get(key: string): string | null {
    if (typeof window !== "undefined") {
      return Cookies.get(key) || null;
    }
    return null;
  }
  set(key: string, value: string, options?: StorageOptions): void {
    if (typeof window !== "undefined") {
      Cookies.set(key, value, { expires: options?.expires ?? 7, path: options?.path ?? "/" });
    }
  }
  remove(key: string, options?: StorageOptions): void {
    if (typeof window !== "undefined") {
      Cookies.remove(key, { path: options?.path ?? "/" });
    }
  }
}

export class MemoryStorage implements StorageAdapter {
  private store: Map<string, string> = new Map();

  get(key: string): string | null {
    return this.store.get(key) || null;
  }

  set(key: string, value: string): void {
    this.store.set(key, value);
  }

  remove(key: string): void {
    this.store.delete(key);
  }

  clear(): void {
    this.store.clear();
  }
}

export class IdentityClient {
  private client: AxiosInstance;
  private storage: StorageAdapter;
  private accessTokenKey: string;
  private refreshTokenKey: string;

  constructor(
    client: AxiosInstance,
    storage: StorageAdapter,
    accessTokenKey = "auth_token",
    refreshTokenKey = "refresh_token"
  ) {
    this.client = client;
    this.storage = storage;
    this.accessTokenKey = accessTokenKey;
    this.refreshTokenKey = refreshTokenKey;
  }

  // Auth
  async login(email: string, password?: string): Promise<{ token: string; refreshToken?: string }> {
    const response = await this.client.post<{
      data: { access_token: string; refresh_token?: string };
    }>("/api/v1/auth/login", { email, password });

    const { access_token, refresh_token } = response.data.data;
    if (access_token) {
      await this.storage.set(this.accessTokenKey, access_token);
    }
    if (refresh_token) {
      await this.storage.set(this.refreshTokenKey, refresh_token);
    }

    return { token: access_token, refreshToken: refresh_token };
  }

  async register(email: string, password?: string): Promise<{ token: string; refreshToken?: string }> {
    const response = await this.client.post<{
      data: { access_token: string; refresh_token?: string };
    }>("/api/v1/auth/register", { email, password });

    const { access_token, refresh_token } = response.data.data;
    if (access_token) {
      await this.storage.set(this.accessTokenKey, access_token);
    }
    if (refresh_token) {
      await this.storage.set(this.refreshTokenKey, refresh_token);
    }

    return { token: access_token, refreshToken: refresh_token };
  }

  async refreshToken(): Promise<{ token: string; refreshToken?: string }> {
    const refreshToken = await this.storage.get(this.refreshTokenKey);
    if (!refreshToken) {
      throw new Error("No refresh token available");
    }

    const response = await this.client.post<{
      data: { access_token: string; refresh_token?: string };
    }>(
      "/api/v1/auth/refresh",
      { refresh_token: refreshToken },
      { _skipAuthRefresh: true } as any
    );

    const { access_token, refresh_token: newRefreshToken } = response.data.data;
    if (access_token) {
      await this.storage.set(this.accessTokenKey, access_token);
    }
    if (newRefreshToken) {
      await this.storage.set(this.refreshTokenKey, newRefreshToken);
    }

    return { token: access_token, refreshToken: newRefreshToken || refreshToken };
  }

  async logout(): Promise<void> {
    try {
      await this.client.post("/api/v1/auth/logout");
    } catch (e) {
      // Ignore errors on logout (e.g. already logged out)
    } finally {
      await this.clearTokens();
    }
  }

  async clearTokens(): Promise<void> {
    await Promise.all([
      this.storage.remove(this.accessTokenKey),
      this.storage.remove(this.refreshTokenKey),
    ]);
  }

  async getAccessToken(): Promise<string | null> {
    return (await this.storage.get(this.accessTokenKey)) || null;
  }

  async getRefreshToken(): Promise<string | null> {
    return (await this.storage.get(this.refreshTokenKey)) || null;
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
  accessTokenKey?: string;
  refreshTokenKey?: string;
  onAuthExpired?: () => void;
}

function setHeader(config: any, name: string, value: string) {
  if (config.headers && typeof config.headers.set === "function") {
    config.headers.set(name, value);
  } else {
    config.headers = config.headers || {};
    config.headers[name] = value;
  }
}

function deleteHeader(config: any, name: string) {
  if (config.headers && typeof config.headers.delete === "function") {
    config.headers.delete(name);
  } else if (config.headers) {
    delete config.headers[name];
    delete config.headers[name.toLowerCase()];
  }
}

export class HextaSDK {
  public identity: IdentityClient;
  public client: AxiosInstance;
  private storage: StorageAdapter;
  private accessTokenKey: string;
  private refreshTokenKey: string;
  private onAuthExpiredCallback?: () => void;

  private isRefreshing = false;
  private failedQueue: Array<{
    resolve: (token: string) => void;
    reject: (error: any) => void;
  }> = [];

  constructor(config: SDKConfig) {
    this.storage = config.storage || new DefaultBrowserStorage();
    this.accessTokenKey = config.accessTokenKey || "auth_token";
    this.refreshTokenKey = config.refreshTokenKey || "refresh_token";
    this.onAuthExpiredCallback = config.onAuthExpired;

    this.client = axios.create({
      baseURL: config.apiUrl,
      headers: {
        "Content-Type": "application/json",
      },
    });

    // Request interceptor to inject access token
    this.client.interceptors.request.use(async (reqConfig) => {
      if ((reqConfig as any)._skipAuthRefresh) {
        deleteHeader(reqConfig, "Authorization");
        return reqConfig;
      }
      const token = await this.storage.get(this.accessTokenKey);
      if (token) {
        setHeader(reqConfig, "Authorization", `Bearer ${token}`);
      }
      return reqConfig;
    });

    // Response interceptor to handle silent token refresh and API errors
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        const status = error.response ? error.response.status : null;
        if (status !== 401 || !originalRequest) {
          return this.handleApiError(error);
        }

        // Avoid infinite loop if already retried or flagged to skip refresh
        if (originalRequest._retry || (originalRequest as any)._skipAuthRefresh) {
          return this.handleApiError(error);
        }

        // Do not intercept auth endpoints
        const url = originalRequest.url || "";
        if (
          url.includes("/api/v1/auth/login") ||
          url.includes("/api/v1/auth/register") ||
          url.includes("/api/v1/auth/refresh")
        ) {
          return this.handleApiError(error);
        }

        // If refresh is already in progress, enqueue this request
        if (this.isRefreshing) {
          return new Promise<string>((resolve, reject) => {
            this.failedQueue.push({ resolve, reject });
          })
            .then((token) => {
              originalRequest._retry = true;
              setHeader(originalRequest, "Authorization", `Bearer ${token}`);
              return this.client(originalRequest);
            })
            .catch((err) => {
              return Promise.reject(err);
            });
        }

        originalRequest._retry = true;
        this.isRefreshing = true;

        try {
          const refreshToken = await this.storage.get(this.refreshTokenKey);
          if (!refreshToken) {
            throw new Error("No refresh token available");
          }

          const refreshResponse = await this.client.post<{
            data: { access_token: string; refresh_token?: string };
          }>(
            "/api/v1/auth/refresh",
            { refresh_token: refreshToken },
            {
              _skipAuthRefresh: true,
            } as any
          );

          const responseData = refreshResponse.data?.data;
          const newAccessToken = responseData?.access_token;
          const newRefreshToken = responseData?.refresh_token;

          if (!newAccessToken) {
            throw new Error("Invalid refresh token response");
          }

          await this.storage.set(this.accessTokenKey, newAccessToken);
          if (newRefreshToken) {
            await this.storage.set(this.refreshTokenKey, newRefreshToken);
          }

          this.isRefreshing = false;
          this.processQueue(null, newAccessToken);

          setHeader(originalRequest, "Authorization", `Bearer ${newAccessToken}`);
          return this.client(originalRequest);
        } catch (refreshErr: any) {
          this.isRefreshing = false;

          // Purge tokens on refresh failure
          await this.clearTokens();

          // Reject queued requests
          this.processQueue(refreshErr, null);

          // Trigger auth expiration callback
          if (this.onAuthExpiredCallback) {
            try {
              this.onAuthExpiredCallback();
            } catch (cbErr) {
              // ignore callback error
            }
          }

          return this.handleApiError(refreshErr.response ? refreshErr : error);
        }
      }
    );

    this.identity = new IdentityClient(
      this.client,
      this.storage,
      this.accessTokenKey,
      this.refreshTokenKey
    );
  }

  private processQueue(error: any, token: string | null = null): void {
    this.failedQueue.forEach((prom) => {
      if (error) {
        prom.reject(error);
      } else if (token) {
        prom.resolve(token);
      }
    });
    this.failedQueue = [];
  }

  private handleApiError(error: any): never {
    if (error.response && error.response.data) {
      const apiError = error.response.data;
      const msg = typeof apiError === "string" ? apiError : (apiError.message || apiError.detail || "API Error");
      const err = new Error(msg);
      (err as any).response = error.response;
      (err as any).status = error.response.status;
      (err as any).code = apiError.code;
      throw err;
    }
    throw error;
  }

  public setOnAuthExpired(callback: () => void): void {
    this.onAuthExpiredCallback = callback;
  }

  public async clearTokens(): Promise<void> {
    await Promise.all([
      this.storage.remove(this.accessTokenKey),
      this.storage.remove(this.refreshTokenKey),
    ]);
  }

  public getStorage(): StorageAdapter {
    return this.storage;
  }
}

// Re-export EcommerceHubSDK for backward compatibility
export { HextaSDK as EcommerceHubSDK };
