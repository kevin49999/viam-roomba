<script lang="ts">
	import { getContext } from 'svelte';
	import { BASE_CONTEXT_KEY, type BaseContextValue } from '$lib/contexts/base';

	const baseContext = getContext<BaseContextValue>(BASE_CONTEXT_KEY);
	if (!baseContext) {
		throw new Error('BaseCommands must be used inside BaseProvider');
	}

	const { doCommandMutation, spinMutation } = baseContext;

	let angleDeg = $state(90);
	let degsPerSec = $state(45);

	function start() {
		doCommandMutation.mutate([{ command: 'start' }]);
	}

	function clean() {
		doCommandMutation.mutate([{ command: 'clean' }]);
	}

	function stop() {
		doCommandMutation.mutate([{ command: 'stop' }]);
	}

	function spin() {
		spinMutation.mutate([angleDeg, degsPerSec]);
	}
</script>

<div class="flex flex-col gap-4">
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

	<div class="flex flex-wrap items-end gap-3 rounded-lg border border-slate-200 p-4 dark:border-slate-800">
		<div class="flex flex-col gap-1">
			<label for="angle-deg" class="text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
				Angle (deg)
			</label>
			<input
				id="angle-deg"
				type="number"
				bind:value={angleDeg}
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900"
			/>
		</div>

		<div class="flex flex-col gap-1">
			<label for="degs-per-sec" class="text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
				Speed (deg/s)
			</label>
			<input
				id="degs-per-sec"
				type="number"
				bind:value={degsPerSec}
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900"
			/>
		</div>

		<button
			type="button"
			onclick={spin}
			disabled={spinMutation.isPending}
			class="rounded bg-slate-800 px-4 py-1.5 text-sm font-medium text-white shadow-sm transition hover:bg-slate-700 disabled:opacity-50 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-white"
		>
			Spin
		</button>

		{#if spinMutation.isError}
			<p class="text-sm text-red-600 dark:text-red-400">{spinMutation.error?.message}</p>
		{/if}
	</div>
</div>
