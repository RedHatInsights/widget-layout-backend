import { Identity, IdentityHeader, IdentityError } from '../types/identity';
import { logger } from './logger';
import { mcpAuthFailureTotal } from './metrics';

export function parseIdentity(headerValue: string | undefined): Identity {
  if (!headerValue) {
    logger.warn('mcp: Missing x-rh-identity header');
    mcpAuthFailureTotal.inc({ reason: 'missing_header' });
    throw new IdentityError('x-rh-identity header is required');
  }

  try {
    const decoded = Buffer.from(headerValue, 'base64').toString('utf-8');
    const identityHeader: IdentityHeader = JSON.parse(decoded);

    if (!identityHeader.identity) {
      logger.warn('mcp: Invalid identity header structure - missing identity field');
      mcpAuthFailureTotal.inc({ reason: 'invalid_structure' });
      throw new IdentityError('Invalid identity header structure');
    }

    if (!identityHeader.identity.org_id) {
      logger.warn('mcp: Missing org_id in identity');
      mcpAuthFailureTotal.inc({ reason: 'missing_org_id' });
      throw new IdentityError('org_id is required in identity');
    }

    const identity: Identity = {
      org_id: identityHeader.identity.org_id,
      user: identityHeader.identity.user,
      internal: identityHeader.identity.internal,
      rawHeader: headerValue,
    };

    logger.debug({ org_id: identity.org_id }, 'mcp: Parsed identity successfully');
    return identity;
  } catch (error) {
    if (error instanceof IdentityError) {
      throw error;
    }
    logger.error({ error }, 'mcp: Failed to parse identity header');
    mcpAuthFailureTotal.inc({ reason: 'parse_error' });
    throw new IdentityError('Failed to parse x-rh-identity header');
  }
}

export function validateIdentity(identity: Identity | null, requiresAuth: boolean): void {
  if (requiresAuth && !identity) {
    throw new IdentityError('Authentication required for this tool');
  }
}
