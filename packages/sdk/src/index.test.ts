import { describe, it, expect, beforeEach, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { HextaSDK, EcommerceHubSDK, MemoryStorage } from "./index";

describe("HextaSDK and Interceptors", () => {
  let storage: MemoryStorage;
  let sdk: HextaSDK;
  let mock: MockAdapter;

  beforeEach(() => {
    storage = new MemoryStorage();
    sdk = new HextaSDK({
      apiUrl: "http://api.test",
      storage,
    });
    mock = new MockAdapter(sdk.client);
  });

  describe("Branding and Aliases", () => {
    it("exports EcommerceHubSDK as an alias to HextaSDK", () => {
      expect(EcommerceHubSDK).toBe(HextaSDK);
    });

    it("provides MemoryStorage adapter", () => {
      storage.set("key", "val");
      expect(storage.get("key")).toBe("val");
      storage.remove("key");
      expect(storage.get("key")).toBeNull();
    });
  });

  describe("Request Interceptor", () => {
    it("injects Authorization Bearer header when access token is in storage", async () => {
      await storage.set("auth_token", "existing-access-token");
      mock.onGet("/api/v1/test").reply((config) => {
        expect(config.headers?.Authorization).toBe("Bearer existing-access-token");
        return [200, { data: { success: true } }];
      });

      const res = await sdk.client.get("/api/v1/test");
      expect(res.data.data.success).toBe(true);
    });

    it("does not attach Authorization header if no token is stored", async () => {
      mock.onGet("/api/v1/test").reply((config) => {
        expect(config.headers?.Authorization).toBeUndefined();
        return [200, { data: { success: true } }];
      });

      const res = await sdk.client.get("/api/v1/test");
      expect(res.data.data.success).toBe(true);
    });
  });

  describe("IdentityClient Auth Methods", () => {
    it("login saves both access_token and refresh_token", async () => {
      mock.onPost("/api/v1/auth/login").reply(200, {
        data: {
          access_token: "jwt-access-token",
          refresh_token: "jwt-refresh-token",
          session_id: 123,
          user_id: "u-1",
        },
      });

      const result = await sdk.identity.login("user@example.com", "secret123");
      expect(result.token).toBe("jwt-access-token");
      expect(result.refreshToken).toBe("jwt-refresh-token");

      expect(await storage.get("auth_token")).toBe("jwt-access-token");
      expect(await storage.get("refresh_token")).toBe("jwt-refresh-token");
      expect(await sdk.identity.getAccessToken()).toBe("jwt-access-token");
      expect(await sdk.identity.getRefreshToken()).toBe("jwt-refresh-token");
    });

    it("register saves both access_token and refresh_token", async () => {
      mock.onPost("/api/v1/auth/register").reply(200, {
        data: {
          access_token: "jwt-register-access",
          refresh_token: "jwt-register-refresh",
          session_id: 456,
          user_id: "u-2",
        },
      });

      const result = await sdk.identity.register("new@example.com", "secret123");
      expect(result.token).toBe("jwt-register-access");
      expect(result.refreshToken).toBe("jwt-register-refresh");

      expect(await storage.get("auth_token")).toBe("jwt-register-access");
      expect(await storage.get("refresh_token")).toBe("jwt-register-refresh");
    });

    it("refreshToken manually renews tokens via API", async () => {
      await storage.set("refresh_token", "current-refresh-token");
      mock.onPost("/api/v1/auth/refresh").reply((config) => {
        const body = JSON.parse(config.data);
        expect(body.refresh_token).toBe("current-refresh-token");
        return [
          200,
          {
            data: {
              access_token: "new-access-token",
              refresh_token: "new-refresh-token",
            },
          },
        ];
      });

      const res = await sdk.identity.refreshToken();
      expect(res.token).toBe("new-access-token");
      expect(res.refreshToken).toBe("new-refresh-token");
      expect(await storage.get("auth_token")).toBe("new-access-token");
      expect(await storage.get("refresh_token")).toBe("new-refresh-token");
    });

    it("logout purges all stored tokens", async () => {
      await storage.set("auth_token", "access");
      await storage.set("refresh_token", "refresh");
      mock.onPost("/api/v1/auth/logout").reply(200, { message: "Logged out" });

      await sdk.identity.logout();

      expect(await storage.get("auth_token")).toBeNull();
      expect(await storage.get("refresh_token")).toBeNull();
    });
  });

  describe("Silent Token Refresh Interceptor", () => {
    it("automatically refreshes token on 401 and replays original request", async () => {
      await storage.set("auth_token", "expired-access-token");
      await storage.set("refresh_token", "valid-refresh-token");

      let attempts = 0;
      mock.onGet("/api/v1/tenants/t-1").reply((config) => {
        attempts++;
        if (attempts === 1) {
          expect(config.headers?.Authorization).toBe("Bearer expired-access-token");
          return [401, { message: "Token expired" }];
        }
        // Second attempt with refreshed token
        expect(config.headers?.Authorization).toBe("Bearer new-access-token");
        return [200, { data: { id: "t-1", name: "Test Tenant" } }];
      });

      mock.onPost("/api/v1/auth/refresh").reply((config) => {
        const body = JSON.parse(config.data);
        expect(body.refresh_token).toBe("valid-refresh-token");
        return [
          200,
          {
            data: {
              access_token: "new-access-token",
              refresh_token: "rotated-refresh-token",
            },
          },
        ];
      });

      const tenant = await sdk.identity.getTenant("t-1");
      expect(tenant.name).toBe("Test Tenant");
      expect(attempts).toBe(2);
      expect(await storage.get("auth_token")).toBe("new-access-token");
      expect(await storage.get("refresh_token")).toBe("rotated-refresh-token");
    });

    it("queues concurrent requests during token refresh and replays all", async () => {
      await storage.set("auth_token", "expired-token");
      await storage.set("refresh_token", "valid-refresh-token");

      let tenantCalls = 0;
      let userCalls = 0;
      let refreshCalls = 0;

      mock.onGet("/api/v1/tenants/t-1").reply((config) => {
        tenantCalls++;
        if (tenantCalls === 1) {
          return [401, { message: "Token expired" }];
        }
        expect(config.headers?.Authorization).toBe("Bearer brand-new-token");
        return [200, { data: { id: "t-1", name: "Tenant 1" } }];
      });

      mock.onGet("/api/v1/tenants/t-1/users").reply((config) => {
        userCalls++;
        if (userCalls === 1) {
          return [401, { message: "Token expired" }];
        }
        expect(config.headers?.Authorization).toBe("Bearer brand-new-token");
        return [200, { data: [{ id: "u-1", email: "user@test.com" }] }];
      });

      mock.onPost("/api/v1/auth/refresh").reply(() => {
        refreshCalls++;
        return [
          200,
          {
            data: {
              access_token: "brand-new-token",
              refresh_token: "new-refresh-token",
            },
          },
        ];
      });

      // Fire concurrent requests simultaneously
      const [tenant, users] = await Promise.all([
        sdk.identity.getTenant("t-1"),
        sdk.identity.getUsers("t-1"),
      ]);

      expect(tenant.name).toBe("Tenant 1");
      expect(users.length).toBe(1);
      expect(refreshCalls).toBe(1); // Only 1 refresh call initiated
      expect(tenantCalls).toBe(2);
      expect(userCalls).toBe(2);
      expect(await storage.get("auth_token")).toBe("brand-new-token");
    });

    it("purges tokens, rejects requests, and calls onAuthExpired when refresh fails", async () => {
      const onExpired = vi.fn();
      sdk.setOnAuthExpired(onExpired);

      await storage.set("auth_token", "expired-access-token");
      await storage.set("refresh_token", "invalid-refresh-token");

      mock.onGet("/api/v1/tenants/t-1").reply(401, { message: "Unauthorized" });
      mock.onPost("/api/v1/auth/refresh").reply(401, { message: "Refresh token revoked" });

      await expect(sdk.identity.getTenant("t-1")).rejects.toThrow();

      expect(onExpired).toHaveBeenCalledTimes(1);
      expect(await storage.get("auth_token")).toBeNull();
      expect(await storage.get("refresh_token")).toBeNull();
    });

    it("purges tokens and calls onAuthExpired immediately when no refresh token exists", async () => {
      const onExpired = vi.fn();
      sdk.setOnAuthExpired(onExpired);

      await storage.set("auth_token", "expired-access-token");
      // No refresh token in storage

      mock.onGet("/api/v1/tenants/t-1").reply(401, { message: "Unauthorized" });

      await expect(sdk.identity.getTenant("t-1")).rejects.toThrow();

      expect(onExpired).toHaveBeenCalledTimes(1);
      expect(await storage.get("auth_token")).toBeNull();
    });

    it("does not loop on 401 if refresh endpoint itself returns 401", async () => {
      await storage.set("refresh_token", "bad-refresh");
      mock.onPost("/api/v1/auth/refresh").reply(401, { message: "Invalid session" });

      await expect(sdk.identity.refreshToken()).rejects.toThrow();
    });

    it("does not loop if retried request fails with 401 again", async () => {
      await storage.set("auth_token", "token-1");
      await storage.set("refresh_token", "valid-refresh");

      let refreshCount = 0;
      mock.onPost("/api/v1/auth/refresh").reply(() => {
        refreshCount++;
        return [200, { data: { access_token: "token-2" } }];
      });

      // Always return 401 even after retry
      mock.onGet("/api/v1/protected").reply(401, { message: "Permanent Forbidden" });

      await expect(sdk.client.get("/api/v1/protected")).rejects.toThrow();
      expect(refreshCount).toBe(1); // Not retried in an infinite loop
    });
  });

  describe("Tenant Operations", () => {
    it("creates a new tenant", async () => {
      const input = { name: "Acme Corp", slug: "acme-corp", plan: "pro" };
      mock.onPost("/api/v1/tenants").reply((config) => {
        const body = JSON.parse(config.data);
        expect(body.name).toBe("Acme Corp");
        expect(body.slug).toBe("acme-corp");
        return [
          201,
          {
            data: {
              id: "tenant-acme",
              name: "Acme Corp",
              slug: "acme-corp",
              status: "active",
              plan: "pro",
              owner_id: "user-1",
            },
          },
        ];
      });

      const res = await sdk.identity.createTenant(input);
      expect(res.id).toBe("tenant-acme");
      expect(res.name).toBe("Acme Corp");
      expect(res.plan).toBe("pro");
    });

    it("fetches list of tenants", async () => {
      mock.onGet("/api/v1/tenants").reply(200, {
        data: [
          { id: "t-1", name: "Tenant One", slug: "t-1", status: "active", plan: "free", owner_id: "u-1" },
          { id: "t-2", name: "Tenant Two", slug: "t-2", status: "active", plan: "pro", owner_id: "u-2" },
        ],
      });

      const list = await sdk.identity.getTenants();
      expect(list.length).toBe(2);
      expect(list[0].id).toBe("t-1");
      expect(list[1].name).toBe("Tenant Two");
    });

    it("fetches single tenant by ID", async () => {
      mock.onGet("/api/v1/tenants/t-1").reply(200, {
        data: { id: "t-1", name: "Tenant One", slug: "t-1", status: "active", plan: "enterprise", owner_id: "u-1" },
      });

      const tenant = await sdk.identity.getTenant("t-1");
      expect(tenant.id).toBe("t-1");
      expect(tenant.plan).toBe("enterprise");
    });

    it("updates tenant by ID", async () => {
      mock.onPut("/api/v1/tenants/t-1").reply((config) => {
        const body = JSON.parse(config.data);
        expect(body.name).toBe("Updated Tenant");
        return [
          200,
          {
            data: { id: "t-1", name: "Updated Tenant", slug: "t-1", status: "active", plan: "pro", owner_id: "u-1" },
          },
        ];
      });

      const updated = await sdk.identity.updateTenant("t-1", { name: "Updated Tenant" });
      expect(updated.name).toBe("Updated Tenant");
    });
  });

  describe("Member Operations", () => {
    it("fetches team members for a tenant", async () => {
      mock.onGet("/api/v1/tenants/t-1/users").reply(200, {
        data: [
          { id: 1, tenant_id: "t-1", user_id: "u-1", role: "owner" },
          { id: 2, tenant_id: "t-1", user_id: "u-2", role: "admin" },
        ],
      });

      const users = await sdk.identity.getUsers("t-1");
      expect(users.length).toBe(2);
      expect(users[0].role).toBe("owner");
      expect(users[1].user_id).toBe("u-2");
    });

    it("invites a member to a tenant", async () => {
      mock.onPost("/api/v1/tenants/t-1/invites").reply((config) => {
        const body = JSON.parse(config.data);
        expect(body.user_id).toBe("u-3");
        expect(body.role).toBe("member");
        return [
          201,
          {
            data: { id: 3, tenant_id: "t-1", user_id: "u-3", role: "member" },
          },
        ];
      });

      const invite = await sdk.identity.inviteMember("t-1", { user_id: "u-3", role: "member" });
      expect(invite.id).toBe(3);
      expect(invite.user_id).toBe("u-3");
      expect(invite.role).toBe("member");
    });
  });

  describe("Storage and Configuration", () => {
    it("exposes configured storage adapter via getStorage", () => {
      expect(sdk.getStorage()).toBe(storage);
    });

    it("clears tokens via sdk.clearTokens()", async () => {
      await storage.set("auth_token", "test-access");
      await storage.set("refresh_token", "test-refresh");

      await sdk.clearTokens();

      expect(await storage.get("auth_token")).toBeNull();
      expect(await storage.get("refresh_token")).toBeNull();
    });

    it("MemoryStorage clear method purges all stored items", () => {
      const mem = new MemoryStorage();
      mem.set("a", "1");
      mem.set("b", "2");
      expect(mem.get("a")).toBe("1");
      mem.clear();
      expect(mem.get("a")).toBeNull();
      expect(mem.get("b")).toBeNull();
    });
  });

  describe("Error Handling", () => {
    it("unwraps API error messages and codes correctly", async () => {
      mock.onGet("/api/v1/error-test").reply(400, {
        message: "Invalid tenant slug",
        code: "INVALID_SLUG",
      });

      try {
        await sdk.client.get("/api/v1/error-test");
        expect.fail("Should have thrown error");
      } catch (err: any) {
        expect(err.message).toBe("Invalid tenant slug");
        expect(err.status).toBe(400);
        expect(err.code).toBe("INVALID_SLUG");
      }
    });

    it("falls back to detail or default message when message is absent", async () => {
      mock.onGet("/api/v1/error-detail").reply(422, {
        detail: "Unprocessable entity payload",
      });

      await expect(sdk.client.get("/api/v1/error-detail")).rejects.toThrow("Unprocessable entity payload");
    });
  });
});

