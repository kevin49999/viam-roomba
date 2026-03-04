<script lang="ts">
	import { getContext } from 'svelte';
	import { BRAIN_CONTEXT_KEY, type BrainContextValue } from '$lib/contexts/brain';

	const brainContext = getContext<BrainContextValue>(BRAIN_CONTEXT_KEY);
	if (!brainContext) {
		throw new Error('BrainCommands must be used inside BrainProvider');
	}

	const { doCommandMutation } = brainContext;

	function send(command: string, extra?: Record<string, unknown>) {
		doCommandMutation.mutate([{ command, ...extra }]);
	}

	const simpleCommands = [
		{ command: 'stop', label: 'Stop' },
		{ command: 'start', label: 'Start' },
		{ command: 'clean', label: 'Clean' },
		{ command: 'seek_dock', label: 'Seek Dock' },
		{ command: 'reset_position', label: 'Reset Position' },
		{ command: 'add_song', label: 'Add Song' },
	] as const;

	let songNumber = 0;
</script>

<div class="flex flex-col gap-4">
	<div class="flex flex-wrap gap-2">
		{#each simpleCommands as { command, label }}
			<button
				type="button"
				onclick={() => send(command)}
				disabled={doCommandMutation.isPending}
				class="rounded bg-slate-800 px-3 py-1.5 text-sm font-medium text-white shadow-sm transition hover:bg-slate-700 disabled:opacity-50 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-white"
			>
				{label}
			</button>
		{/each}
		<div class="flex items-center gap-2">
			<label for="song-number" class="text-sm font-medium text-slate-700 dark:text-slate-300">Song #</label>
			<input
				id="song-number"
				type="number"
				min="0"
				max="4"
				bind:value={songNumber}
				class="w-16 rounded bg-slate-100 px-2 py-1.5 text-sm text-slate-900 shadow-sm dark:bg-slate-700 dark:text-white"
			/>
			<button
				type="button"
				onclick={() => send('play_song', { song_number: songNumber })}
				disabled={doCommandMutation.isPending}
				class="rounded bg-slate-800 px-3 py-1.5 text-sm font-medium text-white shadow-sm transition hover:bg-slate-700 disabled:opacity-50 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-white"
			>
				Play Song
			</button>
		</div>
	</div>

	{#if doCommandMutation.isError}
		<p class="text-sm text-red-600 dark:text-red-400">{doCommandMutation.error?.message}</p>
	{/if}
	{#if doCommandMutation.isSuccess}
		<p class="text-xs text-slate-500 dark:text-slate-400 font-mono">
			{JSON.stringify(doCommandMutation.data)}
		</p>
	{/if}
</div>
