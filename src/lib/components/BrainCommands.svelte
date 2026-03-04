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
		{ command: 'play_song', label: 'Play LOTR' },
		{ command: 'play_vader_song', label: 'Play Vader' },
	] as const;

	let moveStraightDistanceMm = $state(500);
	let moveStraightMmPerSec = $state(300);

	function moveStraight() {
		doCommandMutation.mutate([{ command: 'move_straight', distance_mm: moveStraightDistanceMm, mm_per_sec: moveStraightMmPerSec }]);
	}
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
	</div>

	<div class="flex flex-wrap items-end gap-3 rounded-lg border border-slate-200 p-4 dark:border-slate-800">
		<div class="flex flex-col gap-1">
			<label for="brain-distance-mm" class="text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
				Distance (mm)
			</label>
			<input
				id="brain-distance-mm"
				type="number"
				bind:value={moveStraightDistanceMm}
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900"
			/>
		</div>
		<div class="flex flex-col gap-1">
			<label for="brain-mm-per-sec" class="text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
				Speed (mm/s)
			</label>
			<input
				id="brain-mm-per-sec"
				type="number"
				bind:value={moveStraightMmPerSec}
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900"
			/>
		</div>
		<button
			type="button"
			onclick={moveStraight}
			disabled={doCommandMutation.isPending}
			class="rounded bg-slate-800 px-4 py-1.5 text-sm font-medium text-white shadow-sm transition hover:bg-slate-700 disabled:opacity-50 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-white"
		>
			Move Straight (Brain)
		</button>
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
