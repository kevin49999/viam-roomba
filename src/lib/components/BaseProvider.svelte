<script lang="ts">
	import { createResourceClient, createResourceMutation } from '@viamrobotics/svelte-sdk';
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

	const contextValue: BaseContextValue = {
		doCommandMutation: doCommandMutation as unknown as BaseContextValue['doCommandMutation'],
		spinMutation: spinMutation as unknown as BaseContextValue['spinMutation'],
		moveStraightMutation: moveStraightMutation as unknown as BaseContextValue['moveStraightMutation'],
		setVelocityMutation: setVelocityMutation as unknown as BaseContextValue['setVelocityMutation']
	};
	setContext(BASE_CONTEXT_KEY, contextValue);
</script>

{#if children}
	{@render children()}
{/if}
