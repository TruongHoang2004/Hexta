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

export class IdentityClient {
  private baseUrl: string;
  private getToken: () => string | null;

  constructor(baseUrl: string, getToken: () => string | null) {
    this.baseUrl = baseUrl;
    this.getToken = getToken;
  }

  private async fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const token = this.getToken();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };

    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        ...headers,
        ...options?.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `API Error: ${response.status}`);
    }

    return response.json();
  }

  // Auth
  async login(email: string, password?: string): Promise<{ token: string }> {
    return this.fetchApi<{ token: string }>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  }

  async register(email: string, password?: string): Promise<{ token: string }> {
    return this.fetchApi<{ token: string }>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  }

  // Tenant
  async getTenant(id: string): Promise<Tenant> {
    return this.fetchApi<Tenant>(`/api/v1/tenants/${id}`);
  }

  // Users
  async getUsers(tenantId: string): Promise<User[]> {
    return this.fetchApi<User[]>(`/api/v1/tenants/${tenantId}/users`);
  }
}

export class EcommerceHubSDK {
  public identity: IdentityClient;

  constructor(config: { apiUrl: string; getToken: () => string | null }) {
    this.identity = new IdentityClient(config.apiUrl, config.getToken);
  }
}
