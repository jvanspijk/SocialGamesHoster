export function hasAssignedRole(
	roleKey: string | undefined,
	roles: ReadonlyArray<{ id: string }>
): boolean {
	const normalizedRoleKey = roleKey?.trim().toLowerCase();
	return (
		normalizedRoleKey !== undefined &&
		normalizedRoleKey !== '' &&
		normalizedRoleKey !== 'unassigned' &&
		roles.some((role) => role.id.toLowerCase() === normalizedRoleKey)
	);
}
