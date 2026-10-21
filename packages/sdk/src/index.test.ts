import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import Cookies from "js-cookie";
import { DefaultBrowserStorage, HextaSDK, EcommerceHubSDK, MemoryStorage, SDKApiError, type StorageAdapter } from "./index";

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>(done => { resolve = done; });
  return { promise, resolve };
}

describe("HextaSDK", () => {
  let storage: MemoryStorage;
  let sdk: HextaSDK;
  let mock: MockAdapter;

  beforeEach(() => {
    storage = new MemoryStorage();
    sdk = new HextaSDK({ apiUrl: "http://api.test", storage });
    mock = new MockAdapter(sdk.client);
  });

  afterEach(() => {
    mock.restore();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("keeps the compatibility alias and configured storage", () => {
    expect(EcommerceHubSDK).toBe(HextaSDK);
    expect(sdk.getStorage()).toBe(storage);
    storage.set("key", "val");
    expect(storage.get("key")).toBe("val");
    storage.clear();
    expect(storage.get("key")).toBeNull();
  });

  it("attaches the access token and includes cookies on requests", async () => {
    storage.set("auth_token", "access");
    mock.onGet("/protected").reply(config => {
      expect(config.headers?.Authorization).toBe("Bearer access");
      expect(config.withCredentials).toBe(true);
      return [200, {}];
    });
    await sdk.client.get("/protected");
  });

  it("leaves Authorization unset when access storage is empty", async () => {
    mock.onGet("/protected").reply(config => {
      expect(config.headers?.Authorization).toBeUndefined();
      return [200, {}];
    });
    await sdk.client.get("/protected");
  });

  describe("server-owned refresh cookie contract", () => {
    it.each(["login", "register"] as const)("%s saves only the access token", async action => {
      const set = vi.spyOn(storage, "set");
      const get = vi.spyOn(storage, "get");
      const remove = vi.spyOn(storage, "remove");
      mock.onPost(`/api/v1/auth/${action}`).reply(config => {
        expect(config.withCredentials).toBe(true);
        expect(JSON.parse(config.data)).toEqual({ email: "user@example.test", password: "password" });
        // Even an older server response must not expose refresh credentials to callers/storage.
        return [200, { data: { access_token: "access", refresh_token: "legacy-credential" } }];
      });
      expect(await sdk.identity[action]("user@example.test", "password")).toEqual({ token: "access" });
      expect(set.mock.calls).toEqual([["auth_token", "access"]]);
      expect(get.mock.calls.every(([key]) => key === "auth_token")).toBe(true);
      expect(remove).not.toHaveBeenCalled();
    });

    it("refreshes with cookies, no token body, and no expired bearer header", async () => {
      storage.set("auth_token", "expired");
      const get = vi.spyOn(storage, "get");
      const set = vi.spyOn(storage, "set");
      mock.onPost("/api/v1/auth/refresh").reply(config => {
        expect(config.withCredentials).toBe(true);
        expect(config.data).toBeUndefined();
        expect(config.headers?.Authorization).toBeUndefined();
        return [200, { data: { access_token: "renewed" } }];
      });
      expect(await sdk.identity.refreshToken()).toEqual({ token: "renewed" });
      expect(set.mock.calls).toEqual([["auth_token", "renewed"]]);
      expect(get.mock.calls.every(([key]) => key === "auth_token")).toBe(true);
    });

    it("logs out through the API and only clears local access state", async () => {
      storage.set("auth_token", "access");
      const remove = vi.spyOn(storage, "remove");
      mock.onPost("/api/v1/auth/logout").reply(config => {
        expect(config.withCredentials).toBe(true);
        expect(config.headers?.Authorization).toBe("Bearer access");
        return [200, {}];
      });
      await sdk.identity.logout();
      expect(await sdk.identity.getAccessToken()).toBeNull();
      expect(remove.mock.calls).toEqual([["auth_token"]]);
    });

    it.each(["server", "network"] as const)("reports %s logout failure and preserves retryable access state", async failure => {
      storage.set("auth_token", "access");
      if (failure === "server") mock.onPost("/api/v1/auth/logout").replyOnce(503, { message: "Unavailable" });
      else mock.onPost("/api/v1/auth/logout").networkErrorOnce();
      await expect(sdk.identity.logout()).rejects.toThrow();
      expect(storage.get("auth_token")).toBe("access");
      mock.onPost("/api/v1/auth/logout").reply(200, {});
      await sdk.identity.logout();
      expect(storage.get("auth_token")).toBeNull();
    });

    it("refreshes expired access once to complete authenticated logout", async () => {
      storage.set("auth_token", "expired");
      mock.onPost("/api/v1/auth/logout").reply(config => {
        return config.headers?.Authorization === "Bearer renewed"
          ? [200, {}] : [401, { message: "Access expired" }];
      });
      mock.onPost("/api/v1/auth/refresh").reply(200, { data: { access_token: "renewed" } });
      await sdk.identity.logout();
      expect(mock.history.post.filter(request => request.url === "/api/v1/auth/logout")).toHaveLength(2);
      expect(mock.history.post.filter(request => request.url === "/api/v1/auth/refresh")).toHaveLength(1);
      expect(storage.get("auth_token")).toBeNull();
      await expect(sdk.identity.refreshToken()).rejects.toThrow("signed out");
      expect(mock.history.post).toHaveLength(3);
    });

    it("fences a refresh response arriving after confirmed server logout", async () => {
      storage.set("auth_token", "access");
      const onExpired = vi.fn();
      sdk.setOnAuthExpired(onExpired);
      const refresh = deferred<[number, { data: { access_token: string } }]>();
      mock.onPost("/api/v1/auth/refresh").reply(() => refresh.promise);
      mock.onPost("/api/v1/auth/logout").reply(200, {});
      const refreshed = sdk.identity.refreshToken().catch((error: unknown) => error);
      await vi.waitFor(() => expect(mock.history.post).toHaveLength(1));
      const logout = sdk.identity.logout();
      await vi.waitFor(() => expect(mock.history.post).toHaveLength(2));
      refresh.resolve([200, { data: { access_token: "obsolete" } }]);
      await logout;
      expect(await refreshed).toBeInstanceOf(Error);
      expect(storage.get("auth_token")).toBeNull();
      expect(onExpired).not.toHaveBeenCalled();
      await expect(sdk.identity.refreshToken()).rejects.toThrow("signed out");
      expect(mock.history.post).toHaveLength(2);
    });

    it("drains an asynchronous access write before successful logout completes", async () => {
      storage.set("auth_token", "access");
      const writeStarted = deferred<void>();
      const finishWrite = deferred<void>();
      const delayedStorage: StorageAdapter = {
        get: key => storage.get(key),
        remove: key => storage.remove(key),
        set: async (key, value) => {
          writeStarted.resolve();
          await finishWrite.promise;
          storage.set(key, value);
        },
      };
      mock.restore();
      sdk = new HextaSDK({ apiUrl: "http://api.test", storage: delayedStorage });
      mock = new MockAdapter(sdk.client);
      mock.onPost("/api/v1/auth/refresh").reply(200, { data: { access_token: "obsolete" } });
      mock.onPost("/api/v1/auth/logout").reply(200, {});
      const refreshed = sdk.identity.refreshToken().catch((error: unknown) => error);
      await writeStarted.promise;
      const logout = sdk.identity.logout();
      await vi.waitFor(() => expect(mock.history.post).toHaveLength(2));
      finishWrite.resolve();
      await logout;
      expect(await refreshed).toBeInstanceOf(Error);
      expect(storage.get("auth_token")).toBeNull();
    });

    it("permits a new session after successful logout", async () => {
      mock.onPost("/api/v1/auth/logout").reply(200, {});
      await sdk.identity.logout();
      mock.onPost("/api/v1/auth/login").reply(200, { data: { access_token: "new-session" } });
      await sdk.identity.login("user@example.test", "password");
      mock.onPost("/api/v1/auth/refresh").reply(200, { data: { access_token: "renewed-session" } });
      expect(await sdk.identity.refreshToken()).toEqual({ token: "renewed-session" });
    });

    it("does not try to delete the HttpOnly cookie when clearing client state", async () => {
      const remove = vi.spyOn(storage, "remove");
      await sdk.clearTokens();
      expect(remove.mock.calls).toEqual([["auth_token"]]);
    });
  });

  describe("single-flight refresh and retries", () => {
    it("refreshes without JavaScript access to refresh credentials", async () => {
      storage.set("auth_token", "expired");
      mock.onGet("/protected").replyOnce(401, { message: "Expired" });
      mock.onGet("/protected").reply(config => {
        expect(config.headers?.Authorization).toBe("Bearer renewed");
        return [200, { ok: true }];
      });
      mock.onPost("/api/v1/auth/refresh").reply(200, { data: { access_token: "renewed" } });
      expect((await sdk.client.get("/protected")).data).toEqual({ ok: true });
      expect(mock.history.post).toHaveLength(1);
    });

    it("coalesces concurrent 401s and replays every waiting request", async () => {
      storage.set("auth_token", "expired");
      const refresh = deferred<[number, { data: { access_token: string } }]>();
      mock.onGet(/\/protected/).reply(config => config.headers?.Authorization === "Bearer renewed"
        ? [200, { ok: true }] : [401, { message: "Expired" }]);
      mock.onPost("/api/v1/auth/refresh").reply(() => refresh.promise);
      const pending = Promise.all([sdk.client.get("/protected/1"), sdk.client.get("/protected/2")]);
      await vi.waitFor(() => {
        expect(mock.history.get).toHaveLength(2);
        expect(mock.history.post).toHaveLength(1);
      });
      refresh.resolve([200, { data: { access_token: "renewed" } }]);
      expect((await pending).map(result => result.data)).toEqual([{ ok: true }, { ok: true }]);
      expect(mock.history.post).toHaveLength(1);
    });

    it("shares an in-flight manual refresh with automatic retries", async () => {
      storage.set("auth_token", "expired");
      const refresh = deferred<[number, { data: { access_token: string } }]>();
      mock.onPost("/api/v1/auth/refresh").reply(() => refresh.promise);
      const manual = sdk.identity.refreshToken();
      mock.onGet("/protected").replyOnce(401, {});
      mock.onGet("/protected").reply(200, {});
      const protectedRequest = sdk.client.get("/protected");
      await vi.waitFor(() => expect(mock.history.get).toHaveLength(1));
      refresh.resolve([200, { data: { access_token: "renewed" } }]);
      await Promise.all([manual, protectedRequest]);
      expect(mock.history.post).toHaveLength(1);
    });

    it("retries a delayed stale-token 401 without another rotation", async () => {
      storage.set("auth_token", "expired");
      const delayed = deferred<[number, object]>();
      mock.onGet("/slow").replyOnce(() => delayed.promise);
      mock.onGet("/slow").reply(200, {});
      const slow = sdk.client.get("/slow");
      mock.onGet("/fast").replyOnce(401, {});
      mock.onGet("/fast").reply(200, {});
      mock.onPost("/api/v1/auth/refresh").reply(200, { data: { access_token: "renewed" } });
      await sdk.client.get("/fast");
      delayed.resolve([401, {}]);
      await slow;
      expect(mock.history.post).toHaveLength(1);
      expect(mock.history.get.at(-1)?.headers?.Authorization).toBe("Bearer renewed");
    });

    it("rejects all waiting requests and reports expiration once when refresh fails", async () => {
      storage.set("auth_token", "expired");
      const onExpired = vi.fn();
      sdk.setOnAuthExpired(onExpired);
      const refresh = deferred<[number, { message: string }]>();
      mock.onPost("/api/v1/auth/refresh").reply(() => refresh.promise);
      mock.onGet(/\/protected/).reply(401, { message: "Expired" });
      const pending = Promise.allSettled([sdk.client.get("/protected/1"), sdk.client.get("/protected/2")]);
      await vi.waitFor(() => expect(mock.history.get).toHaveLength(2));
      refresh.resolve([401, { message: "Refresh expired" }]);
      const results = await pending;
      expect(results.every(result => result.status === "rejected")).toBe(true);
      expect(onExpired).toHaveBeenCalledTimes(1);
      expect(storage.get("auth_token")).toBeNull();
      expect(mock.history.post).toHaveLength(1);
    });

    it("clears access state when refresh returns an invalid payload", async () => {
      storage.set("auth_token", "expired");
      const onExpired = vi.fn(() => { throw new Error("Consumer failure"); });
      sdk.setOnAuthExpired(onExpired);
      mock.onPost("/api/v1/auth/refresh").reply(200, { data: {} });
      await expect(sdk.identity.refreshToken()).rejects.toThrow("Invalid access token response");
      expect(storage.get("auth_token")).toBeNull();
      expect(onExpired).toHaveBeenCalledTimes(1);
    });

    it("never refreshes recursively on an auth failure or a second 401", async () => {
      mock.onPost("/api/v1/auth/login").reply(401, { message: "Bad credentials" });
      await expect(sdk.identity.login("user@example.test", "wrong")).rejects.toThrow("Bad credentials");
      expect(mock.history.post).toHaveLength(1);
      mock.resetHistory();
      mock.onGet("/protected").reply(401, {});
      mock.onPost("/api/v1/auth/refresh").reply(200, { data: { access_token: "renewed" } });
      await expect(sdk.client.get("/protected")).rejects.toThrow();
      expect(mock.history.post).toHaveLength(1);
      expect(mock.history.get).toHaveLength(2);
    });
  });

  describe("Tenant Operations", () => {
    it("requests a tenant without activating it", async () => {
      const input = { name: "Acme Corp", slug: "acme-corp" };
      mock.onPost("/api/v1/tenant-requests").reply(config => {
        expect(JSON.parse(config.data)).toEqual(input);
        return [200, { data: { id: "request-acme", ...input, status: "pending", tenant_id: null } }];
      });
      const result = await sdk.identity.requestTenant(input);
      expect(result.status).toBe("pending");
      expect(result.tenant_id).toBeNull();
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

    it("creates an email invitation without an active membership", async () => {
      mock.onPost("/api/v1/tenants/t-1/invitations").reply(config => {
        expect(JSON.parse(config.data)).toEqual({ email: "employee@example.test" });
        return [200, { data: { invitation: { id: "invite-1", tenant_id: "t-1", email: "employee@example.test", role: "member", status: "pending" }, token: "one-time-token", delivery: "manual_link_not_emailed" } }];
      });
      const result = await sdk.identity.createEmailInvitation("t-1", "employee@example.test");
      expect(result.invitation.status).toBe("pending");
      expect(result.delivery).toBe("manual_link_not_emailed");
      expect(result).not.toHaveProperty("user_id");
    });
  });

  describe("typed API errors", () => {
    it("preserves the error message, status, code, and response", async () => {
      mock.onGet("/error").reply(400, { message: "Invalid slug", code: "INVALID_SLUG" });
      try {
        await sdk.client.get("/error");
        expect.fail("Expected the API error");
      } catch (error: unknown) {
        expect(error).toBeInstanceOf(SDKApiError);
        if (!(error instanceof SDKApiError)) throw error;
        expect(error.message).toBe("Invalid slug");
        expect(error.status).toBe(400);
        expect(error.code).toBe("INVALID_SLUG");
        expect(error.response.status).toBe(400);
      }
    });

    it.each([
      [{ detail: "Invalid payload" }, "Invalid payload"],
      [{ detail: { unexpected: true }, message: "Readable message" }, "Readable message"],
      [{ code: 7 }, "API Error"],
      ["Plain text error", "Plain text error"],
    ])("handles unknown response payloads safely", async (payload, message) => {
      mock.onGet("/error").reply(422, payload);
      await expect(sdk.client.get("/error")).rejects.toThrow(message);
    });
  });

  describe("browser storage options", () => {
    it("forwards security, domain, and path options on set and remove", () => {
      vi.stubGlobal("window", { location: { protocol: "https:" } });
      const set = vi.spyOn(Cookies, "set").mockImplementation(() => undefined);
      const remove = vi.spyOn(Cookies, "remove").mockImplementation(() => undefined);
      const options = { secure: true, sameSite: "strict" as const, domain: "example.test", path: "/app", expires: 1 };
      const browser = new DefaultBrowserStorage();
      browser.set("auth_token", "access", options);
      browser.remove("auth_token", options);
      expect(set).toHaveBeenCalledWith("auth_token", "access", options);
      expect(remove).toHaveBeenCalledWith("auth_token", options);
    });

    it.each([["https:", true], ["http:", false]])("uses safe access-cookie defaults for %s", (protocol, secure) => {
      vi.stubGlobal("window", { location: { protocol } });
      const set = vi.spyOn(Cookies, "set").mockImplementation(() => undefined);
      new DefaultBrowserStorage().set("auth_token", "access");
      expect(set).toHaveBeenCalledWith("auth_token", "access", { expires: 7, path: "/", sameSite: "lax", secure });
    });
  });
});
