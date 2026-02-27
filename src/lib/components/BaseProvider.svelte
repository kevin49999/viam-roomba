<script lang="ts">
	import { createResourceClient, createResourceMutation, createResourceQuery } from '@viamrobotics/svelte-sdk';
	import { BaseClient } from '@viamrobotics/sdk';
	import { setContext } from 'svelte';
	import { BASE_CONTEXT_KEY, type BaseContextValue } from '$lib/contexts/base';

	interface Props {
		partID: string;
		name: string;
		children?: import('svelte').Snippet;
	}

	let { partID, name, children }: Props = $props();

	const baseClient = createResourceClient(BaseClient, () => partID, () => name);
	const doCommandMutation = createResourceMutation(baseClient, 'doCommand');
	const spinMutation = createResourceMutation(baseClient, 'spin');
	const moveStraightMutation = createResourceMutation(baseClient, 'moveStraight');
	const setVelocityMutation = createResourceMutation(baseClient, 'setVelocity');

	const automaticModeQuery = createResourceQuery(baseClient, 'doCommand', () => [{ command: 'get_automatic_mode' }], () => ({
		refetchInterval: 5000
	}));

	let maxLinearSpeed = $state(300);
	let maxAngularSpeed = $state(90);

	const contextValue: BaseContextValue = {
		doCommandMutation: doCommandMutation as unknown as BaseContextValue['doCommandMutation'],
		spinMutation: spinMutation as unknown as BaseContextValue['spinMutation'],
		moveStraightMutation: moveStraightMutation as unknown as BaseContextValue['moveStraightMutation'],
		setVelocityMutation: setVelocityMutation as unknown as BaseContextValue['setVelocityMutation'],
		get maxLinearSpeed() {
			return maxLinearSpeed;
		},
		get maxAngularSpeed() {
			return maxAngularSpeed;
		},
		setMaxLinearSpeed: (speed: number) => {
			maxLinearSpeed = Math.min(Math.max(0, speed), 500); // 500 mm/s is a reasonable absolute max
		},
		setMaxAngularSpeed: (speed: number) => {
			maxAngularSpeed = Math.min(Math.max(0, speed), 180); // 180 deg/s is a reasonable absolute max
		},
		automaticModeQuery: automaticModeQuery as unknown as BaseContextValue['automaticModeQuery'],
		toggleAutomaticMode: async () => {
			await doCommandMutation.mutateAsync([{ command: 'toggle_automatic_mode' }]);
			automaticModeQuery.refetch();
		}
	};
	setContext(BASE_CONTEXT_KEY, contextValue);
</script>

{#if children}
	{@render children()}
{/if}
