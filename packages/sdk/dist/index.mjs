// src/index.ts
import axios from "axios";
import Cookies from "js-cookie";
var DefaultBrowserStorage = class {
  get(key) {
    if (typeof window !== "undefined") {
      return Cookies.get(key) || null;
    }
    return null;
  }
  set(key, value) {
    if (typeof window !== "undefined") {
      Cookies.set(key, value, { expires: 7, path: "/" });
    }
  }
  remove(key) {
    if (typeof window !== "undefined") {
      Cookies.remove(key, { path: "/" });
    }
  }
};
var IdentityClient = class {
  client;
  storage;
  authKey = "auth_token";
  constructor(client, storage) {
    this.client = client;
    this.storage = storage;
  }
  // Auth
  async login(email, password) {
    const response = await this.client.post("/api/v1/auth/login", { email, password });
    const token = response.data.data.access_token;
    if (token) {
      await this.storage.set(this.authKey, token);
    }
    return { token };
  }
  async register(email, password) {
    const response = await this.client.post("/api/v1/auth/register", { email, password });
    const token = response.data.data.access_token;
    if (token) {
      await this.storage.set(this.authKey, token);
    }
    return { token };
  }
  async logout() {
    try {
      await this.client.post("/api/v1/auth/logout");
    } catch (e) {
    } finally {
      await this.storage.remove(this.authKey);
    }
  }
  // Tenant
  async getTenant(id) {
    const response = await this.client.get(`/api/v1/tenants/${id}`);
    return response.data.data;
  }
  // Users
  async getUsers(tenantId) {
    const response = await this.client.get(`/api/v1/tenants/${tenantId}/users`);
    return response.data.data;
  }
};
var EcommerceHubSDK = class {
  identity;
  client;
  storage;
  constructor(config) {
    this.storage = config.storage || new DefaultBrowserStorage();
    this.client = axios.create({
      baseURL: config.apiUrl,
      headers: {
        "Content-Type": "application/json"
      }
    });
    this.client.interceptors.request.use(async (reqConfig) => {
      const token = await this.storage.get("auth_token");
      if (token) {
        reqConfig.headers = reqConfig.headers || {};
        reqConfig.headers.Authorization = `Bearer ${token}`;
      }
      return reqConfig;
    });
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
};
export {
  DefaultBrowserStorage,
  EcommerceHubSDK,
  IdentityClient
};
//# sourceMappingURL=index.mjs.map