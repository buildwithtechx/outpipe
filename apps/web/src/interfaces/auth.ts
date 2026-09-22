export type UserStatus = 'active' | 'disabled';

export type OAuthProvider = 'google' | 'github';

import type { Entity } from '#/interfaces/api';

export type AuthUser = Entity & {
  id: string;
  email: string;
  name: string;
  status: UserStatus;
  emailVerifiedAt?: string;
  lastLoginAt?: string;
};

export type AuthSession = {
  id: string;
  userId: string;
  expiresAt: string;
  lastSeenAt?: string;
  isPlatformAdmin: boolean;
  user: AuthUser;
};

export type OAuthStart = {
  provider: OAuthProvider;
  redirectUri?: string;
};
