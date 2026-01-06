// lib/query/keys.ts

export const queryKeys = {
  // Auth
  auth: {
    root: ["auth"] as const,
    session: () => [...queryKeys.auth.root, "session"] as const,
    user: () => [...queryKeys.auth.root, "user"] as const,
  },

  // Credentials
  credentials: {
    root: ["credentials"] as const,
    list: (filters?: { status?: string }) =>
      [...queryKeys.credentials.root, "list", filters] as const,
    detail: (id: string) =>
      [...queryKeys.credentials.root, "detail", id] as const,
  },

  // Integrations
  integrations: {
    root: ["integrations"] as const,
    list: () => [...queryKeys.integrations.root, "list"] as const,
    connections: () =>
      [...queryKeys.integrations.root, "connections"] as const,
    connection: (providerId: string) =>
      [...queryKeys.integrations.root, "connection", providerId] as const,
  },

  // Profile
  profile: {
    root: ["profile"] as const,
    me: () => [...queryKeys.profile.root, "me"] as const,
    public: (username: string) =>
      [...queryKeys.profile.root, "public", username] as const,
  },
} as const;