<script lang="ts">
	import { getContext } from 'svelte';
	import { BASE_CONTEXT_KEY, type BaseContextValue } from '$lib/contexts/base';

	const baseContext = getContext<BaseContextValue>(BASE_CONTEXT_KEY);
	if (!baseContext) {
		throw new Error('BaseCommands must be used inside BaseProvider');
	}

	const { doCommandMutation } = baseContext;

	function start() {
		doCommandMutation.mutate([{ command: 'start' }]);
	}

	function clean() {
		doCommandMutation.mutate([{ command: 'clean' }]);
	}

	function stop() {
		doCommandMutation.mutate([{ command: 'stop' }]);
	}
</script>

<div class="flex flex-wrap items-center gap-2">
	<button
		type="button"
		onclick={start}
		disabled={doCommandMutation.isPending}
		class="rounded-lg bg-blue-600 px-4 py-2 font-medium text-white shadow-sm transition hover:bg-blue-500 disabled:opacity-50 disabled:pointer-events-none"
	>
		Start
	</button>

	<button
		type="button"
		onclick={clean}
		disabled={doCommandMutation.isPending}
		class="rounded-lg bg-emerald-600 px-4 py-2 font-medium text-white shadow-sm transition hover:bg-emerald-500 disabled:opacity-50 disabled:pointer-events-none"
	>
		Clean
	</button>

	<button
		type="button"
		onclick={stop}
		disabled={doCommandMutation.isPending}
		class="rounded-lg bg-red-600 px-4 py-2 font-medium text-white shadow-sm transition hover:bg-red-500 disabled:opacity-50 disabled:pointer-events-none"
	>
		Stop
	</button>
	
	{#if doCommandMutation.isError}
		<p class="text-sm text-red-600 dark:text-red-400">{doCommandMutation.error?.message}</p>
	{/if}
</div>
