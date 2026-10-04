import { AxiosInstance } from 'axios';

interface Tenant {
    id: string;
    name: string;
    slug: string;
    status: string;
    plan: string;
    owner_id: string;
    created_at?: string;
    updated_at?: string;
}
interface TenantMember {
    id: number;
    tenant_id: string;
    user_id: string;
    role: string;
    created_at?: string;
}
interface User {
    id: string;
    tenant_id: string;
    email?: string;
    full_name?: string;
    status?: string;
    role_id?: string;
    role?: string;
    user_id?: string;
}
interface CreateTenantInput {
    name: string;
    slug: string;
    plan?: string;
}
interface InviteMemberInput {
    user_id: string;
    role?: string;
}
interface StorageAdapter {
    get: (key: string) => string | null | Promise<string | null>;
    set: (key: string, value: string) => void | Promise<void>;
    remove: (key: string) => void | Promise<void>;
}
declare class DefaultBrowserStorage implements StorageAdapter {
    get(key: string): string | null;
    set(key: string, value: string): void;
    remove(key: string): void;
}
declare class IdentityClient {
    private client;
    private storage;
    private authKey;
    constructor(client: AxiosInstance, storage: StorageAdapter);
    login(email: string, password?: string): Promise<{
        token: string;
    }>;
    register(email: string, password?: string): Promise<{
        token: string;
    }>;
    logout(): Promise<void>;
    createTenant(input: CreateTenantInput): Promise<Tenant>;
    getTenants(): Promise<Tenant[]>;
    getTenant(id: string): Promise<Tenant>;
    updateTenant(id: string, input: Partial<CreateTenantInput> & {
        status?: string;
    }): Promise<Tenant>;
    getUsers(tenantId: string): Promise<TenantMember[]>;
    inviteMember(tenantId: string, input: InviteMemberInput): Promise<TenantMember>;
}
interface SDKConfig {
    apiUrl: string;
    storage?: StorageAdapter;
}
declare class EcommerceHubSDK {
    identity: IdentityClient;
    private client;
    private storage;
    constructor(config: SDKConfig);
}

export { type CreateTenantInput, DefaultBrowserStorage, EcommerceHubSDK, IdentityClient, type InviteMemberInput, type SDKConfig, type StorageAdapter, type Tenant, type TenantMember, type User };
