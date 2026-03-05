<script lang="ts">
	import { createResourceClient, createResourceMutation, createResourceQuery } from '@viamrobotics/svelte-sdk';
	import { GenericServiceClient } from '@viamrobotics/sdk';
	import { setContext } from 'svelte';
	import { BRAIN_CONTEXT_KEY, type BrainContextValue } from '$lib/contexts/brain';

	interface Props {
		partID: string;
		name: string;
		children?: import('svelte').Snippet;
	}

	let { partID, name, children }: Props = $props();

	const brainClient = createResourceClient(GenericServiceClient, () => partID, () => name);
	const doCommandMutation = createResourceMutation(brainClient, 'doCommand');

	const automaticModeQuery = createResourceQuery(
		brainClient,
		'doCommand',
		() => [{ command: 'get_automatic_mode' }] as [Record<string, any>],
		() => ({ refetchInterval: 5000 })
	);

	const COMMAND_TIMEOUT_MS = 5000;
	let commandTimedOut = $state(false);

	function runMutate(args: [Record<string, unknown>]) {
		commandTimedOut = false;
		const timeoutId = setTimeout(() => {
			commandTimedOut = true;
		}, COMMAND_TIMEOUT_MS);
		doCommandMutation.mutate(args as any, {
			onSettled: () => clearTimeout(timeoutId)
		});
	}

	async function runMutateAsync(args: [Record<string, unknown>]) {
		commandTimedOut = false;
		const timeoutId = setTimeout(() => {
			commandTimedOut = true;
		}, COMMAND_TIMEOUT_MS);
		try {
			return await doCommandMutation.mutateAsync(args as any);
		} finally {
			clearTimeout(timeoutId);
		}
	}

	const timedDoCommandMutation = {
		get isPending() { return !commandTimedOut && doCommandMutation.isPending; },
		get isError() { return commandTimedOut || doCommandMutation.isError; },
		get isSuccess() { return !commandTimedOut && doCommandMutation.isSuccess; },
		get data() { return doCommandMutation.data; },
		get error() {
			if (commandTimedOut) return new Error('Timed out waiting for robot to respond');
			return doCommandMutation.error;
		},
		mutate: (args: any) => runMutate(args),
		mutateAsync: (args: any) => runMutateAsync(args)
	};

	const contextValue: BrainContextValue = {
		doCommandMutation: timedDoCommandMutation as unknown as BrainContextValue['doCommandMutation'],
		automaticModeQuery: automaticModeQuery as unknown as BrainContextValue['automaticModeQuery'],
		toggleAutomaticMode: async () => {
			await runMutateAsync([{ command: 'toggle_automatic_mode' }]);
			automaticModeQuery.refetch();
		}
	};
	setContext(BRAIN_CONTEXT_KEY, contextValue);
</script>

{#if children}
	{@render children()}
{/if}
