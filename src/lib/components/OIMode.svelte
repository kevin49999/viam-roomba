<script lang="ts">
	import { getContext } from 'svelte';
	import { READINGS_CONTEXT_KEY, type ReadingsContextValue } from '$lib/contexts/readings';
	import { BRAIN_CONTEXT_KEY, type BrainContextValue } from '$lib/contexts/brain';

	const readingsContext = getContext<ReadingsContextValue>(READINGS_CONTEXT_KEY);
	const brainContext = getContext<BrainContextValue>(BRAIN_CONTEXT_KEY);

	if (!readingsContext) {
		throw new Error('OIMode must be used inside ReadingsProvider');
	}
	if (!brainContext) {
		throw new Error('OIMode must be used inside BrainProvider');
	}

	const { query } = readingsContext;
	const { doCommandMutation, automaticModeQuery } = brainContext;

	const isAutomatic = $derived.by(() => {
		const data = automaticModeQuery.data;
		if (data && typeof data === 'object' && !Array.isArray(data)) {
			return (data as Record<string, unknown>)['automatic_mode'] === true;
		}
		return false;
	});

	/** Current OI mode from readings: 'off', 'passive', 'safe', or 'full' */
	const currentMode = $derived.by(() => {
		const data = query.data;
		if (data == null || typeof data !== 'object' || Array.isArray(data)) return undefined;
		return (data as Record<string, unknown>)['oi_mode'] as string | undefined;
	});

	const modes = [
		{ id: 'passive', label: 'Passive', command: 'enter_passive_mode' },
		{ id: 'safe', label: 'Safe', command: 'enter_safe_mode' },
		{ id: 'full', label: 'Full', command: 'enter_full_mode' }
	] as const;

	function setMode(command: string) {
		doCommandMutation.mutate([{ command }]);
	}
</script>

<div class="inline-flex rounded-lg border border-slate-200 bg-slate-50 p-1 dark:border-slate-700 dark:bg-slate-800/50">
	{#each modes as { id, label, command }}
		<button
			type="button"
			onclick={() => setMode(command)}
			disabled={doCommandMutation.isPending || isAutomatic}
			class="relative flex items-center justify-center rounded-md px-4 py-1.5 text-sm font-medium transition-all duration-200
				{currentMode === id 
					? 'bg-white text-blue-600 shadow-sm ring-1 ring-slate-200 dark:bg-slate-700 dark:text-blue-400 dark:ring-slate-600' 
					: 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200'}
				disabled:opacity-50 disabled:pointer-events-none"
			aria-pressed={currentMode === id}
		>
			{label}
			{#if currentMode === id}
				<span class="absolute -bottom-1 left-1/2 h-1 w-1 -translate-x-1/2 rounded-full bg-blue-600 dark:bg-blue-400"></span>
			{/if}
		</button>
	{/each}
</div>

{#if doCommandMutation.isError}
	<p class="mt-1 text-xs text-red-600 dark:text-red-400">{doCommandMutation.error?.message}</p>
{/if}
