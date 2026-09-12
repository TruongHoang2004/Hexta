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
interface TokenPair {
    access_token: string;
    refresh_token?: string;
}
interface StorageOptions {
    expires?: number | Date;
    path?: string;
    domain?: string;
    secure?: boolean;
    sameSite?: "strict" | "Strict" | "lax" | "Lax" | "none" | "None";
}
interface StorageAdapter {
    get: (key: string) => string | null | Promise<string | null>;
    set: (key: string, value: string, options?: StorageOptions) => void | Promise<void>;
    remove: (key: string, options?: StorageOptions) => void | Promise<void>;
}
declare class DefaultBrowserStorage implements StorageAdapter {
    get(key: string): string | null;
    set(key: string, value: string, options?: StorageOptions): void;
    remove(key: string, options?: StorageOptions): void;
}
declare class MemoryStorage implements StorageAdapter {
    private store;
    get(key: string): string | null;
    set(key: string, value: string): void;
    remove(key: string): void;
    clear(): void;
}
declare class IdentityClient {
    private client;
    private storage;
    private accessTokenKey;
    private refreshTokenKey;
    constructor(client: AxiosInstance, storage: StorageAdapter, accessTokenKey?: string, refreshTokenKey?: string);
    login(email: string, password?: string): Promise<{
        token: string;
        refreshToken?: string;
    }>;
    register(email: string, password?: string): Promise<{
        token: string;
        refreshToken?: string;
    }>;
    refreshToken(): Promise<{
        token: string;
        refreshToken?: string;
    }>;
    logout(): Promise<void>;
    clearTokens(): Promise<void>;
    getAccessToken(): Promise<string | null>;
    getRefreshToken(): Promise<string | null>;
    getTenant(id: string): Promise<Tenant>;
    getUsers(tenantId: string): Promise<User[]>;
}
interface SDKConfig {
    apiUrl: string;
    storage?: StorageAdapter;
    accessTokenKey?: string;
    refreshTokenKey?: string;
    onAuthExpired?: () => void;
}
declare class HextaSDK {
    identity: IdentityClient;
    client: AxiosInstance;
    private storage;
    private accessTokenKey;
    private refreshTokenKey;
    private onAuthExpiredCallback?;
    private isRefreshing;
    private failedQueue;
    constructor(config: SDKConfig);
    private processQueue;
    private handleApiError;
    setOnAuthExpired(callback: () => void): void;
    clearTokens(): Promise<void>;
    getStorage(): StorageAdapter;
}

export { DefaultBrowserStorage, HextaSDK as EcommerceHubSDK, HextaSDK, IdentityClient, MemoryStorage, type SDKConfig, type StorageAdapter, type StorageOptions, type Tenant, type TokenPair, type User };
