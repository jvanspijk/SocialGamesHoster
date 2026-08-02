<script lang="ts">
	import CheckboxField from '$lib/components/CheckboxField.svelte';
	import CheckboxGroup from '$lib/components/CheckboxGroup.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField from '$lib/components/SelectField.svelte';

	let audience = $state<'all' | 'team'>('all');
	let visible = $state(false);
	let roles = $state<string[]>([]);
	let tags = $state('initial');

	const audienceOptions = [
		{ value: 'all' as const, label: 'All players' },
		{ value: 'team' as const, label: 'A selected team' }
	];
	const roleOptions = [
		{ value: 'seer', label: 'Seer' },
		{ value: 'wolf', label: 'Wolf' }
	];
</script>

<SelectField
	label="Audience"
	name="audience"
	bind:value={audience}
	options={audienceOptions}
	help="Who hears this cue."
	required
/>
<SelectField
	label="Unavailable choice"
	name="unavailable-choice"
	value="all"
	options={audienceOptions}
	error="Choose an audience."
	disabled
/>
<CheckboxField
	label="Visible to players"
	name="visible"
	bind:checked={visible}
	description="Players can see this channel."
	required
/>
<CheckboxField label="Disabled setting" name="disabled-setting" checked={false} disabled />
<CheckboxField label="Repeated setting" name="repeated-setting" checked={false} />
<CheckboxField label="Repeated setting" name="repeated-setting" checked={false} />
<CheckboxGroup
	label="Roles"
	name="roles"
	bind:values={roles}
	options={roleOptions}
	help="Choose any number of roles."
	required
/>
<CheckboxGroup
	label="Partly disabled roles"
	name="partly-disabled-roles"
	values={[]}
	options={[
		{ value: 'seer', label: 'Disabled seer', disabled: true },
		{ value: 'wolf', label: 'Enabled wolf' }
	]}
	required
/>
<CheckboxGroup
	label="Disabled roles"
	name="disabled-roles"
	values={[]}
	options={roleOptions}
	help="This group is unavailable."
	error="Roles cannot be changed."
	disabled
/>
<SelectField
	label="Repeated audience"
	name="repeated-audience"
	value="all"
	options={audienceOptions}
/>
<SelectField
	label="Repeated audience"
	name="repeated-audience"
	value="all"
	options={audienceOptions}
/>
<CheckboxGroup label="Repeated roles" name="repeated-roles" values={[]} options={roleOptions} />
<CheckboxGroup label="Repeated roles" name="repeated-roles" values={[]} options={roleOptions} />
<Field label="Tags" name="tags" value={tags} onchange={(value) => (tags = value)} />

<output data-testid="audience">{audience}</output>
<output data-testid="visible">{String(visible)}</output>
<output data-testid="roles">{roles.join(',')}</output>
<output data-testid="tags">{tags}</output>
