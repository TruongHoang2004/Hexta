interface Tenant {
    id: string;
    name: string;
    slug: string;
    status: string;
    plan: string;
    owner_id: string;
}
interface User {
    id: string;
    tenant_id: string;
    email: string;
    full_name: string;
    status: string;
    role_id: string;
}
declare class IdentityClient {
    private baseUrl;
    private getToken;
    constructor(baseUrl: string, getToken: () => string | null);
    private fetchApi;
    login(email: string, password?: string): Promise<{
        token: string;
    }>;
    getTenant(id: string): Promise<Tenant>;
    getUsers(tenantId: string): Promise<User[]>;
}
declare class EcommerceHubSDK {
    identity: IdentityClient;
    constructor(config: {
        apiUrl: string;
        getToken: () => string | null;
    });
}

export { EcommerceHubSDK, IdentityClient, type Tenant, type User };
