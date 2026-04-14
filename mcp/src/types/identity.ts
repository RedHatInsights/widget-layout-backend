// x-rh-identity header types

export interface Identity {
  org_id: string;
  user?: {
    username: string;
    email?: string;
    first_name?: string;
    last_name?: string;
    is_active?: boolean;
    is_org_admin?: boolean;
    is_internal?: boolean;
    locale?: string;
    user_id?: string;
  };
  internal?: {
    org_id?: string;
    auth_type?: string;
  };
  rawHeader: string;
}

export interface IdentityHeader {
  identity: {
    account_number?: string;
    org_id: string;
    type: string;
    user?: {
      username: string;
      email?: string;
      first_name?: string;
      last_name?: string;
      is_active?: boolean;
      is_org_admin?: boolean;
      is_internal?: boolean;
      locale?: string;
      user_id?: string;
    };
    internal?: {
      org_id?: string;
      auth_type?: string;
    };
  };
}

export class IdentityError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'IdentityError';
  }
}
