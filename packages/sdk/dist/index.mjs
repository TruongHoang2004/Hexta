// src/index.ts
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
export {
  EcommerceHubSDK,
  IdentityClient
};
//# sourceMappingURL=index.mjs.map