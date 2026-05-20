import { cleanEmptyFields } from './object.utils';

describe('cleanEmptyFields', () => {
  it('should remove empty strings', () => {
    const input = { name: 'John', email: '', phone: '123' };
    const result = cleanEmptyFields(input);
    expect(result).toEqual({ name: 'John', phone: '123' });
  });

  it('should remove null values', () => {
    const input = { name: 'John', email: null, phone: '123' };
    const result = cleanEmptyFields(input);
    expect(result).toEqual({ name: 'John', phone: '123' });
  });

  it('should remove undefined values', () => {
    const input = { name: 'John', email: undefined, phone: '123' };
    const result = cleanEmptyFields(input);
    expect(result).toEqual({ name: 'John', phone: '123' });
  });

  it('should keep zero values', () => {
    const input = { name: 'John', count: 0, phone: '123' };
    const result = cleanEmptyFields(input);
    expect(result).toEqual({ name: 'John', count: 0, phone: '123' });
  });

  it('should keep false values', () => {
    const input = { name: 'John', active: false, phone: '123' };
    const result = cleanEmptyFields(input);
    expect(result).toEqual({ name: 'John', active: false, phone: '123' });
  });

  it('should handle empty object', () => {
    const input = {};
    const result = cleanEmptyFields(input);
    expect(result).toEqual({});
  });

  it('should handle object with all empty values', () => {
    const input = { email: '', phone: null, address: undefined };
    const result = cleanEmptyFields(input);
    expect(result).toEqual({});
  });

  it('should handle object with all valid values', () => {
    const input = { name: 'John', email: 'john@example.com', phone: '123' };
    const result = cleanEmptyFields(input);
    expect(result).toEqual(input);
  });

  it('should preserve nested objects', () => {
    const input = {
      name: 'John',
      address: { street: '123 Main', city: 'NYC' },
      phone: '',
    };
    const result = cleanEmptyFields(input);
    expect(result).toEqual({
      name: 'John',
      address: { street: '123 Main', city: 'NYC' },
    });
  });

  it('should preserve arrays', () => {
    const input = {
      name: 'John',
      tags: ['tag1', 'tag2'],
      phone: '',
    };
    const result = cleanEmptyFields(input);
    expect(result).toEqual({
      name: 'John',
      tags: ['tag1', 'tag2'],
    });
  });

  it('should preserve empty arrays', () => {
    const input = {
      name: 'John',
      tags: [],
      phone: '',
    };
    const result = cleanEmptyFields(input);
    expect(result).toEqual({
      name: 'John',
      tags: [],
    });
  });

  it('should not mutate original object', () => {
    const input = { name: 'John', email: '', phone: '123' };
    const original = { ...input };
    cleanEmptyFields(input);
    expect(input).toEqual(original);
  });

  it('should handle complex form object', () => {
    const formValue = {
      name: 'Jane Doe',
      tenant_type: 'Individual',
      nid_number: '123456',
      phone_number: '8801234567',
      email: '',
      address: 'Dhaka',
    };
    const result = cleanEmptyFields(formValue);
    expect(result).toEqual({
      name: 'Jane Doe',
      tenant_type: 'Individual',
      nid_number: '123456',
      phone_number: '8801234567',
      address: 'Dhaka',
    });
  });

  it('should handle whitespace-only strings', () => {
    const input = { name: '  ', email: 'test@example.com' };
    const result = cleanEmptyFields(input);
    expect(result).toEqual({ name: '  ', email: 'test@example.com' });
  });
});
