import { AxiosInstance, AxiosResponse } from 'axios';

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
interface TenantApplicationInput {
    name: string;
    slug: string;
}
interface TenantApplication {
    id: string;
    name: string;
    slug: string;
    status: "pending" | "approved" | "rejected";
    tenant_id: string | null;
}
interface PendingTenantInvitation {
    invitation: {
        id: string;
        tenant_id: string;
        email: string;
        role: string;
        status: string;
        expires_at: string;
    };
    token: string;
    delivery: string;
}
interface AccessTokenResponse {
    access_token: string;
}
type TokenPair = AccessTokenResponse;
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
    private onAuthExpired?;
    private refreshPromise;
    private logoutPromise;
    private authGeneration;
    private signedOut;
    constructor(client: AxiosInstance, storage: StorageAdapter, accessTokenKey?: string, onAuthExpired?: (() => void) | undefined);
    login(email: string, password?: string): Promise<{
        token: string;
    }>;
    register(email: string, password?: string): Promise<{
        token: string;
    }>;
    refreshToken(): Promise<{
        token: string;
    }>;
    private renewAccessToken;
    private saveAccessToken;
    logout(): Promise<void>;
    private revokeSession;
    clearTokens(): Promise<void>;
    getAccessToken(): Promise<string | null>;
    requestTenant(input: TenantApplicationInput): Promise<TenantApplication>;
    getTenants(): Promise<Tenant[]>;
    getTenant(id: string): Promise<Tenant>;
    updateTenant(id: string, input: {
        name?: string;
    }): Promise<Tenant>;
    getUsers(tenantId: string): Promise<TenantMember[]>;
    createEmailInvitation(tenantId: string, email: string): Promise<PendingTenantInvitation>;
}
interface SDKConfig {
    apiUrl: string;
    storage?: StorageAdapter;
    accessTokenKey?: string;
    onAuthExpired?: () => void;
}
declare class SDKApiError extends Error {
    readonly response: AxiosResponse<unknown>;
    readonly status: number;
    readonly code?: string | undefined;
    constructor(message: string, response: AxiosResponse<unknown>, status: number, code?: string | undefined);
}
declare class HextaSDK {
    identity: IdentityClient;
    client: AxiosInstance;
    private storage;
    private accessTokenKey;
    private onAuthExpiredCallback?;
    constructor(config: SDKConfig);
    private handleApiError;
    setOnAuthExpired(callback: () => void): void;
    clearTokens(): Promise<void>;
    getStorage(): StorageAdapter;
}

export { type AccessTokenResponse, DefaultBrowserStorage, HextaSDK as EcommerceHubSDK, HextaSDK, IdentityClient, MemoryStorage, type PendingTenantInvitation, SDKApiError, type SDKConfig, type StorageAdapter, type StorageOptions, type Tenant, type TenantApplication, type TenantApplicationInput, type TenantMember, type TokenPair, type User };
