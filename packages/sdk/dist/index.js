"use strict";
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
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
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/index.ts
var index_exports = {};
__export(index_exports, {
  EcommerceHubSDK: () => EcommerceHubSDK,
  IdentityClient: () => IdentityClient
});
module.exports = __toCommonJS(index_exports);
var IdentityClient = class {
  baseUrl;
  getToken;
  constructor(baseUrl, getToken) {
    this.baseUrl = baseUrl;
    this.getToken = getToken;
  }
  async fetchApi(endpoint, options) {
    const token = this.getToken();
    const headers = {
      "Content-Type": "application/json"
    };
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        ...headers,
        ...options?.headers
      }
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `API Error: ${response.status}`);
    }
    return response.json();
  }
  // Auth
  async login(email, password) {
    return this.fetchApi("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password })
    });
  }
  async register(email, password) {
    return this.fetchApi("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password })
    });
  }
  // Tenant
  async getTenant(id) {
    return this.fetchApi(`/api/v1/tenants/${id}`);
  }
  // Users
  async getUsers(tenantId) {
    return this.fetchApi(`/api/v1/tenants/${tenantId}/users`);
  }
};
var EcommerceHubSDK = class {
  identity;
  constructor(config) {
    this.identity = new IdentityClient(config.apiUrl, config.getToken);
  }
};
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  EcommerceHubSDK,
  IdentityClient
});
//# sourceMappingURL=index.js.map