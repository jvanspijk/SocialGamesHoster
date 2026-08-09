<script lang="ts">
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { PlayerGameView, RulesetAbility } from '$lib/api/types';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';
	import RoleRevealView, {
		type AbilityAction,
		type RoleAbilityPresentation,
		type RoleKnowledgePresentation,
		type RoleRevealPresentation
	} from './RoleRevealView.svelte';

	let {
		view,
		revealed,
		reveal,
		hide,
		back
	}: {
		view: PlayerGameView;
		revealed: boolean;
		reveal: () => void;
		hide: () => void;
		back: () => void;
	} = $props();

	const role = $derived(view.role);
	const rolePresentation = $derived(
		role
			? ({
					name: role.name,
					description: role.description,
					teamName: role.team?.name,
					winCondition: role.winCondition
				} satisfies RoleRevealPresentation)
			: null
	);
	const roleAsset = $derived(
		role?.imageAssetKey
			? view.assets.find((asset) => asset.kind === 'image' && asset.assetKey === role.imageAssetKey)
			: undefined
	);
	const knowledge = $derived(
		view.knowledge.map((item, index) => ({
			id: `${item.participantId ?? item.seatNumber ?? index}`,
			text: knowledgeText(item)
		})) satisfies RoleKnowledgePresentation[]
	);
	const abilities = $derived(role?.abilities.map(abilityPresentation) ?? []);
	let abilityBusy = $state('');

	function choiceFor(abilityID: string) {
		return (view.abilityChoices ?? []).find((choice) => choice.abilityId === abilityID);
	}

	function abilityPresentation(ability: RulesetAbility): RoleAbilityPresentation {
		const choice = choiceFor(ability.id);
		const activationPhase = (ability.activationPhaseIds ?? []).includes(view.game.phaseKey);
		if (choice) {
			return {
				id: ability.id,
				name: ability.name,
				description: ability.description,
				status: { label: choice.status, tone: choice.status === 'Finalized' ? 'success' : 'info' },
				action:
					choice.status === 'Activated'
						? { label: 'Undo activation', command: 'undo', variant: 'secondary' }
						: undefined
			};
		}

		if (
			activationPhase &&
			!view.game.abilityPhaseLockedAt &&
			view.participant.status === 'active'
		) {
			return {
				id: ability.id,
				name: ability.name,
				description: ability.description,
				action: { label: 'Activate', command: 'activate' }
			};
		}

		return {
			id: ability.id,
			name: ability.name,
			description: ability.description,
			status:
				view.game.abilityPhaseLockedAt && activationPhase
					? { label: 'Choices finalized', tone: 'success' }
					: undefined,
			unavailableLabel:
				view.game.abilityPhaseLockedAt && activationPhase ? undefined : 'Not playable in this phase'
		};
	}

	async function runAbilityAction(abilityID: string, action: AbilityAction) {
		abilityBusy = abilityID;
		const undo = action === 'undo';
		try {
			await api(
				`/games/${view.game.id}/abilities/${abilityID}/activate`,
				undo ? { method: 'DELETE' } : { method: 'POST', ...jsonBody({}) }
			);
			await gameState.refreshPlayer();
			toasts.success(undo ? 'Ability activation undone.' : 'Ability activated.');
		} catch (caught) {
			toasts.error(
				errorMessage(
					caught,
					undo ? 'The activation could not be undone.' : 'The ability could not be activated.'
				)
			);
		} finally {
			abilityBusy = '';
		}
	}

	function knowledgeText(item: Record<string, unknown>) {
		const name = String(item.displayName ?? `Seat ${item.seatNumber ?? ''}`).trim();
		if (item.role && typeof item.role === 'object') {
			const roleName = String((item.role as { name?: unknown }).name ?? 'role known');
			return `${name}: ${roleName}`;
		}
		if (item.teamId) return `${name}: team ${String(item.teamId)}`;
		if (item.status) return `${name}: ${String(item.status)}`;
		return name;
	}
</script>

<RoleRevealView
	available={view.roleAvailable && rolePresentation !== null}
	role={rolePresentation}
	roleAsset={roleAsset?.preview}
	{knowledge}
	{abilities}
	{revealed}
	busyAbilityId={abilityBusy}
	{reveal}
	{hide}
	{back}
	onAbilityAction={runAbilityAction}
/>
