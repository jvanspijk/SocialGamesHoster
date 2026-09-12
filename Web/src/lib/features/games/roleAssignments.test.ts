import { describe, expect, it } from 'vitest';
import { hasAssignedRole } from './roleAssignments';

const roles = [{ id: 'lookout' }, { id: 'captain' }];

describe('hasAssignedRole', () => {
	it.each([undefined, '', '  ', 'Unassigned', 'removed-role'])(
		'treats %j as unassigned',
		(roleKey) => {
			expect(hasAssignedRole(roleKey, roles)).toBe(false);
		}
	);

	it('recognizes a role in the current ruleset regardless of key casing', () => {
		expect(hasAssignedRole(' CAPTAIN ', roles)).toBe(true);
	});
});
