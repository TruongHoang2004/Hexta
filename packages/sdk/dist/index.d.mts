import { AxiosInstance } from 'axios';

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
    getTenant(id: string): Promise<Tenant>;
    getUsers(tenantId: string): Promise<User[]>;
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

export { DefaultBrowserStorage, EcommerceHubSDK, IdentityClient, type SDKConfig, type StorageAdapter, type Tenant, type User };
