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

	const contextValue: BaseContextValue = {
		doCommandMutation: doCommandMutation as unknown as BaseContextValue['doCommandMutation']
	};
	setContext(BASE_CONTEXT_KEY, contextValue);
</script>

{#if children}
	{@render children()}
{/if}
