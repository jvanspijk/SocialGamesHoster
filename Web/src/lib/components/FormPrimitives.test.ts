import { cleanup, fireEvent, render, screen, within } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';
import FormPrimitivesHarness from '../../test/FormPrimitivesHarness.svelte';
import CheckboxFieldSource from './CheckboxField.svelte?raw';
import SelectFieldSource from './SelectField.svelte?raw';

describe('native form primitives', () => {
	afterEach(cleanup);

	it('associates labels, help, errors, required state, and disabled state', () => {
		render(FormPrimitivesHarness);

		const audience = screen.getByRole('combobox', { name: /^Audience/ });
		expect(audience).toBeRequired();
		expect(audience).toHaveAccessibleDescription('Who hears this cue.');

		const unavailable = screen.getByRole('combobox', { name: /^Unavailable choice/ });
		expect(unavailable).toBeDisabled();
		expect(unavailable).toHaveAccessibleDescription('Choose an audience.');
		expect(unavailable).toHaveAttribute('aria-invalid', 'true');

		const visible = screen.getByRole('checkbox', {
			name: /^Visible to players/
		}) as HTMLInputElement;
		expect(visible).toBeRequired();
		expect(visible).toHaveAccessibleDescription('Players can see this channel.');
		expect(screen.getByRole('checkbox', { name: 'Disabled setting' })).toBeDisabled();
	});

	it('uses native focusable controls with 44px minimum targets', () => {
		render(FormPrimitivesHarness);

		const audience = screen.getByRole('combobox', { name: /^Audience/ });
		audience.focus();
		expect(audience).toHaveFocus();
		expect(SelectFieldSource).toContain('min-height: var(--target-size)');

		const visible = screen.getByRole('checkbox', {
			name: /^Visible to players/
		}) as HTMLInputElement;
		visible.focus();
		expect(visible).toHaveFocus();
		expect(CheckboxFieldSource).toContain('min-height: var(--target-size)');
	});

	it('forwards edited Field values and creates unique IDs for repeated instances', async () => {
		render(FormPrimitivesHarness);

		await fireEvent.change(screen.getByRole('textbox', { name: 'Tags' }), {
			target: { value: 'investigative, unique' }
		});
		expect(screen.getByTestId('tags')).toHaveTextContent('investigative, unique');

		const repeatedSelects = screen.getAllByRole('combobox', { name: 'Repeated audience' });
		expect(new Set(repeatedSelects.map((select) => select.id)).size).toBe(2);
		const repeatedGroups = screen.getAllByRole('group', { name: 'Repeated roles' });
		expect(
			new Set(
				repeatedGroups.flatMap((group) =>
					Array.from(group.querySelectorAll('input')).map((input) => input.id)
				)
			).size
		).toBe(4);
		const repeatedFields = screen.getAllByRole('checkbox', { name: 'Repeated setting' });
		expect(new Set(repeatedFields.map((field) => field.id)).size).toBe(2);
	});

	it('validates checkbox groups as any selected choice', async () => {
		render(FormPrimitivesHarness);

		const rolesGroup = screen.getByRole('group', { name: 'Roles' });
		const seer = within(rolesGroup).getByRole('checkbox', { name: 'Seer' });
		const wolf = within(rolesGroup).getByRole('checkbox', { name: 'Wolf' });
		expect(seer).not.toBeRequired();
		expect(seer).toBeInvalid();
		await fireEvent.click(wolf);
		expect(seer).toBeValid();

		const partlyDisabled = screen.getByRole('group', { name: 'Partly disabled roles' });
		expect(within(partlyDisabled).getByRole('checkbox', { name: 'Disabled seer' })).toBeDisabled();
		expect(within(partlyDisabled).getByRole('checkbox', { name: 'Enabled wolf' })).toBeInvalid();

		const disabledGroup = screen.getByRole('group', { name: 'Disabled roles' });
		expect(disabledGroup).toHaveAccessibleDescription('Roles cannot be changed.');
		for (const checkbox of Array.from(disabledGroup.querySelectorAll('input'))) {
			expect(checkbox).toBeDisabled();
		}
	});

	it('preserves typed select values and checkbox-group names while bound values update', async () => {
		render(FormPrimitivesHarness);

		const audience = screen.getByRole('combobox', { name: /^Audience/ });
		await fireEvent.change(audience, { target: { value: 'team' } });
		expect(screen.getByTestId('audience')).toHaveTextContent('team');

		const visible = screen.getByRole('checkbox', { name: /^Visible to players/ });
		await fireEvent.click(visible);
		expect(screen.getByTestId('visible')).toHaveTextContent('true');

		const rolesGroup = screen.getByRole('group', { name: 'Roles' });
		const seer = within(rolesGroup).getByRole('checkbox', { name: 'Seer' });
		const wolf = within(rolesGroup).getByRole('checkbox', { name: 'Wolf' });
		expect(seer).toHaveAttribute('name', 'roles');
		expect(wolf).toHaveAttribute('name', 'roles');
		await fireEvent.click(seer);
		await fireEvent.click(wolf);
		expect(screen.getByTestId('roles')).toHaveTextContent('seer,wolf');
	});
});
