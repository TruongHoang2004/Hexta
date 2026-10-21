import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from "axios";
import Cookies from "js-cookie";

export interface Tenant {
  id: string;
  name: string;
  slug: string;
  status: string;
  plan: string;
  owner_id: string;
  created_at?: string;
  updated_at?: string;
}

export interface TenantMember {
  id: number;
  tenant_id: string;
  user_id: string;
  role: string;
  created_at?: string;
}

export interface User {
  id: string;
  tenant_id: string;
  email?: string;
  full_name?: string;
  status?: string;
  role_id?: string;
  role?: string;
  user_id?: string;
}

export interface TenantApplicationInput { name: string; slug: string; }
export interface TenantApplication { id: string; name: string; slug: string; status: "pending" | "approved" | "rejected"; tenant_id: string | null; }
export interface PendingTenantInvitation { invitation: { id: string; tenant_id: string; email: string; role: string; status: string; expires_at: string }; token: string; delivery: string; }

export interface AccessTokenResponse {
  access_token: string;
}

// Kept as a type alias for consumers upgrading from the former token-pair contract.
export type TokenPair = AccessTokenResponse;

interface AuthRequestMetadata {
  _retry?: boolean;
  _skipAuthRefresh?: boolean;
}

type AuthRequestConfig = AxiosRequestConfig<unknown> & AuthRequestMetadata;
type AuthInternalRequestConfig = InternalAxiosRequestConfig<unknown> & AuthRequestMetadata;

export interface StorageOptions {
  expires?: number | Date;
  path?: string;
  domain?: string;
  secure?: boolean;
  sameSite?: "strict" | "Strict" | "lax" | "Lax" | "none" | "None";
}

// Storage adapters hold short-lived access tokens only. Refresh credentials are
// managed by the API in an HttpOnly cookie and must never enter this adapter.
export interface StorageAdapter {
  get: (key: string) => string | null | Promise<string | null>;
  set: (key: string, value: string, options?: StorageOptions) => void | Promise<void>;
  remove: (key: string, options?: StorageOptions) => void | Promise<void>;
}

export class DefaultBrowserStorage implements StorageAdapter {
  get(key: string): string | null {
    return typeof window !== "undefined" ? Cookies.get(key) || null : null;
  }

  set(key: string, value: string, options?: StorageOptions): void {
    if (typeof window !== "undefined") {
      Cookies.set(key, value, {
        ...options,
        expires: options?.expires ?? 7,
        path: options?.path ?? "/",
        sameSite: options?.sameSite ?? "lax",
        secure: options?.secure ?? window.location.protocol === "https:",
      });
    }
  }

  remove(key: string, options?: StorageOptions): void {
    if (typeof window !== "undefined") {
      Cookies.remove(key, { ...options, path: options?.path ?? "/" });
    }
  }
}

export class MemoryStorage implements StorageAdapter {
  private store = new Map<string, string>();

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
  private refreshPromise: Promise<{ token: string }> | null = null;
  private logoutPromise: Promise<void> | null = null;
  private authGeneration = 0;
  private signedOut = false;

  constructor(
    private client: AxiosInstance,
    private storage: StorageAdapter,
    private accessTokenKey = "auth_token",
    private onAuthExpired?: () => void,
  ) {}

  async login(email: string, password?: string): Promise<{ token: string }> {
    const response = await this.client.post<{ data: AccessTokenResponse }>(
      "/api/v1/auth/login", { email, password },
    );
    this.authGeneration++;
    this.signedOut = false;
    return this.saveAccessToken(response.data.data);
  }

  async register(email: string, password?: string): Promise<{ token: string }> {
    const response = await this.client.post<{ data: AccessTokenResponse }>(
      "/api/v1/auth/register", { email, password },
    );
    this.authGeneration++;
    this.signedOut = false;
    return this.saveAccessToken(response.data.data);
  }

  // Manual refresh and interceptor retries share one request so a rotating
  // HttpOnly cookie is never consumed by parallel refresh requests in this SDK.
  refreshToken(): Promise<{ token: string }> {
    if (this.signedOut) return Promise.reject(new Error("The session has been signed out"));
    if (!this.refreshPromise) {
      this.refreshPromise = this.renewAccessToken().finally(() => {
        this.refreshPromise = null;
      });
    }
    return this.refreshPromise;
  }

  private async renewAccessToken(): Promise<{ token: string }> {
    const generation = this.authGeneration;
    try {
      const config: AuthRequestConfig = { _skipAuthRefresh: true };
      const response = await this.client.post<{ data: AccessTokenResponse }>(
        "/api/v1/auth/refresh", undefined, config,
      );
      if (generation !== this.authGeneration) throw new Error("The authentication session changed during refresh");
      const result = await this.saveAccessToken(response.data.data);
      if (generation !== this.authGeneration) throw new Error("The authentication session changed during refresh");
      return result;
    } catch (error: unknown) {
      // Intentional logout or a new login must not expire a newer session.
      if (generation !== this.authGeneration) throw error;
      await this.clearTokens();
      try {
        this.onAuthExpired?.();
      } catch {
        // Consumer callbacks must not replace the authentication error.
      }
      throw error;
    }
  }

  private async saveAccessToken(data: AccessTokenResponse): Promise<{ token: string }> {
    if (typeof data?.access_token !== "string" || !data.access_token) {
      throw new Error("Invalid access token response");
    }
    await this.storage.set(this.accessTokenKey, data.access_token);
    return { token: data.access_token };
  }

  logout(): Promise<void> {
    if (!this.logoutPromise) {
      this.logoutPromise = this.revokeSession().finally(() => {
        this.logoutPromise = null;
      });
    }
    return this.logoutPromise;
  }

  private async revokeSession(): Promise<void> {
    // A failed request leaves local access state available for a retry. Only
    // the server can confirm revocation and expire the HttpOnly refresh cookie.
    await this.client.post("/api/v1/auth/logout");
    this.authGeneration++;
    this.signedOut = true;
    // Fence pending refresh responses, then drain any access-storage write that
    // already started before revocation. The final clear runs after that write.
    await this.refreshPromise?.catch(() => undefined);
    await this.clearTokens();
  }

  // Clears local access state only; use logout() to revoke the server session.
  async clearTokens(): Promise<void> {
    await this.storage.remove(this.accessTokenKey);
  }

  async getAccessToken(): Promise<string | null> {
    return (await this.storage.get(this.accessTokenKey)) || null;
  }

  // Tenant
  async requestTenant(input: TenantApplicationInput): Promise<TenantApplication> {
    const response = await this.client.post<{ data: TenantApplication }>("/api/v1/tenant-requests", input);
    return response.data.data;
  }

  async getTenants(): Promise<Tenant[]> {
    const response = await this.client.get<{ data: Tenant[] }>("/api/v1/tenants");
    return response.data.data;
  }

  async getTenant(id: string): Promise<Tenant> {
    const response = await this.client.get<{ data: Tenant }>(`/api/v1/tenants/${id}`);
    return response.data.data;
  }

  async updateTenant(id: string, input: { name?: string }): Promise<Tenant> {
    const response = await this.client.put<{ data: Tenant }>(`/api/v1/tenants/${id}`, input);
    return response.data.data;
  }

  // Members
  async getUsers(tenantId: string): Promise<TenantMember[]> {
    const response = await this.client.get<{ data: TenantMember[] }>(`/api/v1/tenants/${tenantId}/users`);
    return response.data.data;
  }

  async createEmailInvitation(tenantId: string, email: string): Promise<PendingTenantInvitation> {
    const response = await this.client.post<{ data: PendingTenantInvitation }>(`/api/v1/tenants/${tenantId}/invitations`, { email });
    return response.data.data;
  }

}

export interface SDKConfig {
  apiUrl: string;
  storage?: StorageAdapter;
  accessTokenKey?: string;
  onAuthExpired?: () => void;
}

export class SDKApiError extends Error {
  constructor(
    message: string,
    public readonly response: AxiosResponse<unknown>,
    public readonly status: number,
    public readonly code?: string,
  ) {
    super(message);
    this.name = "SDKApiError";
  }
}

export class HextaSDK {
  public identity: IdentityClient;
  public client: AxiosInstance;
  private storage: StorageAdapter;
  private accessTokenKey: string;
  private onAuthExpiredCallback?: () => void;

  constructor(config: SDKConfig) {
    this.storage = config.storage || new DefaultBrowserStorage();
    this.accessTokenKey = config.accessTokenKey || "auth_token";
    this.onAuthExpiredCallback = config.onAuthExpired;

    this.client = axios.create({
      baseURL: config.apiUrl,
      withCredentials: true,
      headers: { "Content-Type": "application/json" },
    });
    this.identity = new IdentityClient(
      this.client, this.storage, this.accessTokenKey,
      () => this.onAuthExpiredCallback?.(),
    );

    this.client.interceptors.request.use(async (request) => {
      const authRequest = request as AuthInternalRequestConfig;
      if (authRequest._skipAuthRefresh) {
        request.headers.delete("Authorization");
        return request;
      }
      const token = await this.storage.get(this.accessTokenKey);
      if (token) {
        request.headers.set("Authorization", `Bearer ${token}`);
      }
      return request;
    });

    this.client.interceptors.response.use(
      response => response,
      async (error: unknown) => {
        if (!axios.isAxiosError<unknown, unknown>(error)) {
          throw error;
        }
        const original = error.config as AuthInternalRequestConfig | undefined;
        if (error.response?.status !== 401 || !original || original._retry || original._skipAuthRefresh) {
          return this.handleApiError(error);
        }
        const url = original.url || "";
        if (["login", "register", "refresh"].some(action => url.includes(`/api/v1/auth/${action}`))) {
          return this.handleApiError(error);
        }

        original._retry = true;
        const currentToken = await this.storage.get(this.accessTokenKey);
        // A delayed 401 may refer to a token replaced by an earlier refresh.
        // Retry that request with the current token before rotating again.
        if (currentToken && original.headers.get("Authorization") !== `Bearer ${currentToken}`) {
          original.headers.set("Authorization", `Bearer ${currentToken}`);
        } else {
          const { token } = await this.identity.refreshToken();
          original.headers.set("Authorization", `Bearer ${token}`);
        }
        return this.client(original);
      },
    );
  }

  private handleApiError(error: unknown): never {
    if (axios.isAxiosError<unknown, unknown>(error) && error.response) {
      const payload = error.response.data;
      let message = "API Error";
      let code: string | undefined;
      if (typeof payload === "string") {
        message = payload;
      } else if (typeof payload === "object" && payload !== null) {
        const fields = payload as Record<string, unknown>;
        if (typeof fields.detail === "string") message = fields.detail;
        else if (typeof fields.message === "string") message = fields.message;
        if (typeof fields.code === "string") code = fields.code;
      }
      throw new SDKApiError(message, error.response, error.response.status, code);
    }
    throw error;
  }

  public setOnAuthExpired(callback: () => void): void {
    this.onAuthExpiredCallback = callback;
  }

  public async clearTokens(): Promise<void> {
    await this.identity.clearTokens();
  }

  public getStorage(): StorageAdapter {
    return this.storage;
  }
}

// Re-export EcommerceHubSDK for backward compatibility.
export { HextaSDK as EcommerceHubSDK };
