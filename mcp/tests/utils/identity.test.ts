import { parseIdentity, validateIdentity } from '../../src/utils/identity';
import { IdentityError } from '../../src/types/identity';

describe('identity utils', () => {
  describe('parseIdentity', () => {
    it('should parse valid identity header', () => {
      const identityData = {
        identity: {
          org_id: '12345',
          type: 'User',
          user: {
            username: 'testuser',
            email: 'test@example.com',
          },
        },
      };

      const header = Buffer.from(JSON.stringify(identityData)).toString('base64');
      const identity = parseIdentity(header);

      expect(identity.org_id).toBe('12345');
      expect(identity.user?.username).toBe('testuser');
      expect(identity.rawHeader).toBe(header);
    });

    it('should throw error for missing header', () => {
      expect(() => parseIdentity(undefined)).toThrow(IdentityError);
      expect(() => parseIdentity(undefined)).toThrow('x-rh-identity header is required');
    });

    it('should throw error for missing org_id', () => {
      const identityData = {
        identity: {
          type: 'User',
        },
      };

      const header = Buffer.from(JSON.stringify(identityData)).toString('base64');
      expect(() => parseIdentity(header)).toThrow(IdentityError);
      expect(() => parseIdentity(header)).toThrow('org_id is required');
    });

    it('should throw error for invalid JSON', () => {
      const header = Buffer.from('invalid json').toString('base64');
      expect(() => parseIdentity(header)).toThrow(IdentityError);
    });

    it('should throw error for missing identity field', () => {
      const identityData = {
        some_other_field: 'value',
      };

      const header = Buffer.from(JSON.stringify(identityData)).toString('base64');
      expect(() => parseIdentity(header)).toThrow(IdentityError);
      expect(() => parseIdentity(header)).toThrow('Invalid identity header structure');
    });
  });

  describe('validateIdentity', () => {
    it('should not throw for valid identity when auth required', () => {
      const identity = {
        org_id: '12345',
        rawHeader: 'header',
      };

      expect(() => validateIdentity(identity, true)).not.toThrow();
    });

    it('should throw when auth required but no identity', () => {
      expect(() => validateIdentity(null, true)).toThrow(IdentityError);
      expect(() => validateIdentity(null, true)).toThrow('Authentication required');
    });

    it('should not throw when auth not required and no identity', () => {
      expect(() => validateIdentity(null, false)).not.toThrow();
    });
  });
});
