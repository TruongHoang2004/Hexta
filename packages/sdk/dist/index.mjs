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
  set(key, value, options) {
    if (typeof window !== "undefined") {
      Cookies.set(key, value, { expires: options?.expires ?? 7, path: options?.path ?? "/" });
    }
  }
  remove(key, options) {
    if (typeof window !== "undefined") {
      Cookies.remove(key, { path: options?.path ?? "/" });
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
  client;
  storage;
  accessTokenKey;
  refreshTokenKey;
  constructor(client, storage, accessTokenKey = "auth_token", refreshTokenKey = "refresh_token") {
    this.client = client;
    this.storage = storage;
    this.accessTokenKey = accessTokenKey;
    this.refreshTokenKey = refreshTokenKey;
  }
  // Auth
  async login(email, password) {
    const response = await this.client.post("/api/v1/auth/login", { email, password });
    const { access_token, refresh_token } = response.data.data;
    if (access_token) {
      await this.storage.set(this.accessTokenKey, access_token);
    }
    if (refresh_token) {
      await this.storage.set(this.refreshTokenKey, refresh_token);
    }
    return { token: access_token, refreshToken: refresh_token };
  }
  async register(email, password) {
    const response = await this.client.post("/api/v1/auth/register", { email, password });
    const { access_token, refresh_token } = response.data.data;
    if (access_token) {
      await this.storage.set(this.accessTokenKey, access_token);
    }
    if (refresh_token) {
      await this.storage.set(this.refreshTokenKey, refresh_token);
    }
    return { token: access_token, refreshToken: refresh_token };
  }
  async refreshToken() {
    const refreshToken = await this.storage.get(this.refreshTokenKey);
    if (!refreshToken) {
      throw new Error("No refresh token available");
    }
    const response = await this.client.post(
      "/api/v1/auth/refresh",
      { refresh_token: refreshToken },
      { _skipAuthRefresh: true }
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
  async logout() {
    try {
      await this.client.post("/api/v1/auth/logout");
    } catch (e) {
    } finally {
      await this.clearTokens();
    }
  }
  async clearTokens() {
    await Promise.all([
      this.storage.remove(this.accessTokenKey),
      this.storage.remove(this.refreshTokenKey)
    ]);
  }
  async getAccessToken() {
    return await this.storage.get(this.accessTokenKey) || null;
  }
  async getRefreshToken() {
    return await this.storage.get(this.refreshTokenKey) || null;
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
function setHeader(config, name, value) {
  if (config.headers && typeof config.headers.set === "function") {
    config.headers.set(name, value);
  } else {
    config.headers = config.headers || {};
    config.headers[name] = value;
  }
}
function deleteHeader(config, name) {
  if (config.headers && typeof config.headers.delete === "function") {
    config.headers.delete(name);
  } else if (config.headers) {
    delete config.headers[name];
    delete config.headers[name.toLowerCase()];
  }
}
var HextaSDK = class {
  identity;
  client;
  storage;
  accessTokenKey;
  refreshTokenKey;
  onAuthExpiredCallback;
  isRefreshing = false;
  failedQueue = [];
  constructor(config) {
    this.storage = config.storage || new DefaultBrowserStorage();
    this.accessTokenKey = config.accessTokenKey || "auth_token";
    this.refreshTokenKey = config.refreshTokenKey || "refresh_token";
    this.onAuthExpiredCallback = config.onAuthExpired;
    this.client = axios.create({
      baseURL: config.apiUrl,
      headers: {
        "Content-Type": "application/json"
      }
    });
    this.client.interceptors.request.use(async (reqConfig) => {
      if (reqConfig._skipAuthRefresh) {
        deleteHeader(reqConfig, "Authorization");
        return reqConfig;
      }
      const token = await this.storage.get(this.accessTokenKey);
      if (token) {
        setHeader(reqConfig, "Authorization", `Bearer ${token}`);
      }
      return reqConfig;
    });
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;
        const status = error.response ? error.response.status : null;
        if (status !== 401 || !originalRequest) {
          return this.handleApiError(error);
        }
        if (originalRequest._retry || originalRequest._skipAuthRefresh) {
          return this.handleApiError(error);
        }
        const url = originalRequest.url || "";
        if (url.includes("/api/v1/auth/login") || url.includes("/api/v1/auth/register") || url.includes("/api/v1/auth/refresh")) {
          return this.handleApiError(error);
        }
        if (this.isRefreshing) {
          return new Promise((resolve, reject) => {
            this.failedQueue.push({ resolve, reject });
          }).then((token) => {
            originalRequest._retry = true;
            setHeader(originalRequest, "Authorization", `Bearer ${token}`);
            return this.client(originalRequest);
          }).catch((err) => {
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
          const refreshResponse = await this.client.post(
            "/api/v1/auth/refresh",
            { refresh_token: refreshToken },
            {
              _skipAuthRefresh: true
            }
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
        } catch (refreshErr) {
          this.isRefreshing = false;
          await this.clearTokens();
          this.processQueue(refreshErr, null);
          if (this.onAuthExpiredCallback) {
            try {
              this.onAuthExpiredCallback();
            } catch (cbErr) {
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
  processQueue(error, token = null) {
    this.failedQueue.forEach((prom) => {
      if (error) {
        prom.reject(error);
      } else if (token) {
        prom.resolve(token);
      }
    });
    this.failedQueue = [];
  }
  handleApiError(error) {
    if (error.response && error.response.data) {
      const apiError = error.response.data;
      const msg = typeof apiError === "string" ? apiError : apiError.message || apiError.detail || "API Error";
      const err = new Error(msg);
      err.response = error.response;
      err.status = error.response.status;
      err.code = apiError.code;
      throw err;
    }
    throw error;
  }
  setOnAuthExpired(callback) {
    this.onAuthExpiredCallback = callback;
  }
  async clearTokens() {
    await Promise.all([
      this.storage.remove(this.accessTokenKey),
      this.storage.remove(this.refreshTokenKey)
    ]);
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
  MemoryStorage
};
//# sourceMappingURL=index.mjs.map