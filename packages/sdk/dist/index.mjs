// src/index.ts
import axios from "axios";
import Cookies from "js-cookie";
var DefaultBrowserStorage = class {
  get(key) {
    return typeof window !== "undefined" ? Cookies.get(key) || null : null;
  }
  set(key, value, options) {
    if (typeof window !== "undefined") {
      Cookies.set(key, value, {
        ...options,
        expires: options?.expires ?? 7,
        path: options?.path ?? "/",
        sameSite: options?.sameSite ?? "lax",
        secure: options?.secure ?? window.location.protocol === "https:"
      });
    }
  }
  remove(key, options) {
    if (typeof window !== "undefined") {
      Cookies.remove(key, { ...options, path: options?.path ?? "/" });
    }
  }
};
var MemoryStorage = class {
  store = /* @__PURE__ */ new Map();
  get(key) {
    return this.store.get(key) || null;
  }
  set(key, value) {
    this.store.set(key, value);
  }
  remove(key) {
    this.store.delete(key);
  }
  clear() {
    this.store.clear();
  }
};
var IdentityClient = class {
  constructor(client, storage, accessTokenKey = "auth_token", onAuthExpired) {
    this.client = client;
    this.storage = storage;
    this.accessTokenKey = accessTokenKey;
    this.onAuthExpired = onAuthExpired;
  }
  client;
  storage;
  accessTokenKey;
  onAuthExpired;
  refreshPromise = null;
  logoutPromise = null;
  authGeneration = 0;
  signedOut = false;
  async login(email, password) {
    const response = await this.client.post(
      "/api/v1/auth/login",
      { email, password }
    );
    this.authGeneration++;
    this.signedOut = false;
    return this.saveAccessToken(response.data.data);
  }
  async register(email, password) {
    const response = await this.client.post(
      "/api/v1/auth/register",
      { email, password }
    );
    this.authGeneration++;
    this.signedOut = false;
    return this.saveAccessToken(response.data.data);
  }
  // Manual refresh and interceptor retries share one request so a rotating
  // HttpOnly cookie is never consumed by parallel refresh requests in this SDK.
  refreshToken() {
    if (this.signedOut) return Promise.reject(new Error("The session has been signed out"));
    if (!this.refreshPromise) {
      this.refreshPromise = this.renewAccessToken().finally(() => {
        this.refreshPromise = null;
      });
    }
    return this.refreshPromise;
  }
  async renewAccessToken() {
    const generation = this.authGeneration;
    try {
      const config = { _skipAuthRefresh: true };
      const response = await this.client.post(
        "/api/v1/auth/refresh",
        void 0,
        config
      );
      if (generation !== this.authGeneration) throw new Error("The authentication session changed during refresh");
      const result = await this.saveAccessToken(response.data.data);
      if (generation !== this.authGeneration) throw new Error("The authentication session changed during refresh");
      return result;
    } catch (error) {
      if (generation !== this.authGeneration) throw error;
      await this.clearTokens();
      try {
        this.onAuthExpired?.();
      } catch {
      }
      throw error;
    }
  }
  async saveAccessToken(data) {
    if (typeof data?.access_token !== "string" || !data.access_token) {
      throw new Error("Invalid access token response");
    }
    await this.storage.set(this.accessTokenKey, data.access_token);
    return { token: data.access_token };
  }
  logout() {
    if (!this.logoutPromise) {
      this.logoutPromise = this.revokeSession().finally(() => {
        this.logoutPromise = null;
      });
    }
    return this.logoutPromise;
  }
  async revokeSession() {
    await this.client.post("/api/v1/auth/logout");
    this.authGeneration++;
    this.signedOut = true;
    await this.refreshPromise?.catch(() => void 0);
    await this.clearTokens();
  }
  // Clears local access state only; use logout() to revoke the server session.
  async clearTokens() {
    await this.storage.remove(this.accessTokenKey);
  }
  async getAccessToken() {
    return await this.storage.get(this.accessTokenKey) || null;
  }
  // Tenant
  async requestTenant(input) {
    const response = await this.client.post("/api/v1/tenant-requests", input);
    return response.data.data;
  }
  async getTenants() {
    const response = await this.client.get("/api/v1/tenants");
    return response.data.data;
  }
  async getTenant(id) {
    const response = await this.client.get(`/api/v1/tenants/${id}`);
    return response.data.data;
  }
  async updateTenant(id, input) {
    const response = await this.client.put(`/api/v1/tenants/${id}`, input);
    return response.data.data;
  }
  // Members
  async getUsers(tenantId) {
    const response = await this.client.get(`/api/v1/tenants/${tenantId}/users`);
    return response.data.data;
  }
  async createEmailInvitation(tenantId, email) {
    const response = await this.client.post(`/api/v1/tenants/${tenantId}/invitations`, { email });
    return response.data.data;
  }
};
var SDKApiError = class extends Error {
  constructor(message, response, status, code) {
    super(message);
    this.response = response;
    this.status = status;
    this.code = code;
    this.name = "SDKApiError";
  }
  response;
  status;
  code;
};
var HextaSDK = class {
  identity;
  client;
  storage;
  accessTokenKey;
  onAuthExpiredCallback;
  constructor(config) {
    this.storage = config.storage || new DefaultBrowserStorage();
    this.accessTokenKey = config.accessTokenKey || "auth_token";
    this.onAuthExpiredCallback = config.onAuthExpired;
    this.client = axios.create({
      baseURL: config.apiUrl,
      withCredentials: true,
      headers: { "Content-Type": "application/json" }
    });
    this.identity = new IdentityClient(
      this.client,
      this.storage,
      this.accessTokenKey,
      () => this.onAuthExpiredCallback?.()
    );
    this.client.interceptors.request.use(async (request) => {
      const authRequest = request;
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
      (response) => response,
      async (error) => {
        if (!axios.isAxiosError(error)) {
          throw error;
        }
        const original = error.config;
        if (error.response?.status !== 401 || !original || original._retry || original._skipAuthRefresh) {
          return this.handleApiError(error);
        }
        const url = original.url || "";
        if (["login", "register", "refresh"].some((action) => url.includes(`/api/v1/auth/${action}`))) {
          return this.handleApiError(error);
        }
        original._retry = true;
        const currentToken = await this.storage.get(this.accessTokenKey);
        if (currentToken && original.headers.get("Authorization") !== `Bearer ${currentToken}`) {
          original.headers.set("Authorization", `Bearer ${currentToken}`);
        } else {
          const { token } = await this.identity.refreshToken();
          original.headers.set("Authorization", `Bearer ${token}`);
        }
        return this.client(original);
      }
    );
  }
  handleApiError(error) {
    if (axios.isAxiosError(error) && error.response) {
      const payload = error.response.data;
      let message = "API Error";
      let code;
      if (typeof payload === "string") {
        message = payload;
      } else if (typeof payload === "object" && payload !== null) {
        const fields = payload;
        if (typeof fields.detail === "string") message = fields.detail;
        else if (typeof fields.message === "string") message = fields.message;
        if (typeof fields.code === "string") code = fields.code;
      }
      throw new SDKApiError(message, error.response, error.response.status, code);
    }
    throw error;
  }
  setOnAuthExpired(callback) {
    this.onAuthExpiredCallback = callback;
  }
  async clearTokens() {
    await this.identity.clearTokens();
  }
  getStorage() {
    return this.storage;
  }
};
export {
  DefaultBrowserStorage,
  HextaSDK as EcommerceHubSDK,
  HextaSDK,
  IdentityClient,
  MemoryStorage,
  SDKApiError
};
//# sourceMappingURL=index.mjs.map