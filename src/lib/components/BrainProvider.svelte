<script lang="ts">
	import { createResourceClient, createResourceMutation, createResourceQuery } from '@viamrobotics/svelte-sdk';
	import { GenericClient } from '@viamrobotics/sdk';
	import { setContext } from 'svelte';
	import { BRAIN_CONTEXT_KEY, type BrainContextValue } from '$lib/contexts/brain';

	interface Props {
		partID: string;
		name: string;
		children?: import('svelte').Snippet;
	}

	let { partID, name, children }: Props = $props();

	const brainClient = createResourceClient(GenericClient, () => partID, () => name);
	const doCommandMutation = createResourceMutation(brainClient, 'doCommand');

	const automaticModeQuery = createResourceQuery(
		brainClient,
		'doCommand',
		() => [{ command: 'get_automatic_mode' }] as [Record<string, any>],
		() => ({ refetchInterval: 5000 })
	);

	const contextValue: BrainContextValue = {
		doCommandMutation: doCommandMutation as unknown as BrainContextValue['doCommandMutation'],
		automaticModeQuery: automaticModeQuery as unknown as BrainContextValue['automaticModeQuery'],
		toggleAutomaticMode: async () => {
			await doCommandMutation.mutateAsync([{ command: 'toggle_automatic_mode' }]);
			automaticModeQuery.refetch();
		}
	};
	setContext(BRAIN_CONTEXT_KEY, contextValue);
</script>

{#if children}
	{@render children()}
{/if}
