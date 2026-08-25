"use strict";
var __create = Object.create;
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
  // If the importer is in node compatibility mode or this is not an ESM
  // file that has been converted to a CommonJS file using a Babel-
  // compatible transform (i.e. "__esModule" has not been set), then set
  // "default" to the CommonJS "module.exports" for node compatibility.
  isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target,
  mod
));
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/index.ts
var index_exports = {};
__export(index_exports, {
  DefaultBrowserStorage: () => DefaultBrowserStorage,
  EcommerceHubSDK: () => EcommerceHubSDK,
  IdentityClient: () => IdentityClient
});
module.exports = __toCommonJS(index_exports);
var import_axios = __toESM(require("axios"));
var import_js_cookie = __toESM(require("js-cookie"));
var DefaultBrowserStorage = class {
  get(key) {
    if (typeof window !== "undefined") {
      return import_js_cookie.default.get(key) || null;
    }
    return null;
  }
  set(key, value) {
    if (typeof window !== "undefined") {
      import_js_cookie.default.set(key, value, { expires: 7, path: "/" });
    }
  }
  remove(key) {
    if (typeof window !== "undefined") {
      import_js_cookie.default.remove(key, { path: "/" });
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
    this.client = import_axios.default.create({
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
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  DefaultBrowserStorage,
  EcommerceHubSDK,
  IdentityClient
});
//# sourceMappingURL=index.js.map