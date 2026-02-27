<script lang="ts">
	import { getContext } from 'svelte';
	import { BASE_CONTEXT_KEY, type BaseContextValue } from '$lib/contexts/base';

	const baseContext = getContext<BaseContextValue>(BASE_CONTEXT_KEY);
	if (!baseContext) {
		throw new Error('BaseCommands must be used inside BaseProvider');
	}

	const { doCommandMutation, spinMutation, moveStraightMutation, maxLinearSpeed, maxAngularSpeed, setMaxLinearSpeed, setMaxAngularSpeed, automaticModeQuery } = baseContext;

	const isAutomatic = $derived.by(() => {
		const data = automaticModeQuery.data;
		if (data && typeof data === 'object' && !Array.isArray(data)) {
			return (data as Record<string, unknown>)['automatic_mode'] === true;
		}
		return false;
	});

	let angleDeg = $state(90);
	let degsPerSec = $state(45);

	let distanceMm = $state(500);
	let mmPerSec = $state(100);

	function start() {
		doCommandMutation.mutate([{ command: 'start' }]);
	}

	function clean() {
		doCommandMutation.mutate([{ command: 'clean' }]);
	}

	function stop() {
		doCommandMutation.mutate([{ command: 'stop' }]);
	}

	function seekDock() {
		doCommandMutation.mutate([{ command: 'seek_dock' }]);
	}

	function spin() {
		spinMutation.mutate([angleDeg, degsPerSec]);
	}

	function moveStraight() {
		moveStraightMutation.mutate([distanceMm, mmPerSec]);
	}
	function resetPosition() {
		doCommandMutation.mutate([{ command: 'reset_position' }]);
	}

	async function playSong() {
		try {
			await doCommandMutation.mutateAsync([{ command: 'add_song' }]);
			await doCommandMutation.mutateAsync([{ command: 'play_song' }]);
		} catch (error) {
			console.error('Failed to play song:', error);
		}
	}
</script>

<div class="flex flex-col gap-4">
	<div class="flex flex-wrap items-center gap-2">
		<button
			type="button"
			onclick={start}
			disabled={doCommandMutation.isPending || isAutomatic}
			class="rounded-lg bg-blue-600 px-4 py-2 font-medium text-white shadow-sm transition hover:bg-blue-500 disabled:opacity-50 disabled:pointer-events-none"
		>
			Start
		</button>

		<button
			type="button"
			onclick={clean}
			disabled={doCommandMutation.isPending || isAutomatic}
			class="rounded-lg bg-emerald-600 px-4 py-2 font-medium text-white shadow-sm transition hover:bg-emerald-500 disabled:opacity-50 disabled:pointer-events-none"
		>
			Clean
		</button>

		<button
			type="button"
			onclick={stop}
			disabled={doCommandMutation.isPending || isAutomatic}
			class="rounded-lg bg-red-600 px-4 py-2 font-medium text-white shadow-sm transition hover:bg-red-500 disabled:opacity-50 disabled:pointer-events-none"
		>
			Stop
		</button>

		<button
			type="button"
			onclick={seekDock}
			disabled={doCommandMutation.isPending || isAutomatic}
			class="rounded-lg bg-slate-600 px-4 py-2 font-medium text-white shadow-sm transition hover:bg-slate-500 disabled:opacity-50 disabled:pointer-events-none"
		>
			Seek Dock
		</button>

		<button
			type="button"
			onclick={resetPosition}
			disabled={doCommandMutation.isPending || isAutomatic}
			class="rounded-lg bg-slate-600 px-4 py-2 font-medium text-white shadow-sm transition hover:bg-slate-500 disabled:opacity-50 disabled:pointer-events-none"
		>
			Reset Position
		</button>

		<button
			type="button"
			onclick={playSong}
			disabled={doCommandMutation.isPending || isAutomatic}
			class="rounded-lg bg-purple-600 px-4 py-2 font-medium text-white shadow-sm transition hover:bg-purple-500 disabled:opacity-50 disabled:pointer-events-none"
		>
			Play Song
		</button>
		{#if doCommandMutation.isError}
			<p class="text-sm text-red-600 dark:text-red-400">{doCommandMutation.error?.message}</p>
		{/if}
	</div>

	<div class="flex flex-wrap items-end gap-3 rounded-lg border border-slate-200 p-4 dark:border-slate-800">
		<div class="flex flex-col gap-1">
			<label for="max-linear-speed" class="text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
				Max Linear Speed (mm/s)
			</label>
			<input
				id="max-linear-speed"
				type="number"
				value={maxLinearSpeed}
				onchange={(e) => setMaxLinearSpeed(Number((e.target as HTMLInputElement).value))}
				min="0"
				max="500"
				disabled={isAutomatic}
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900 disabled:opacity-50"
			/>
		</div>

		<div class="flex flex-col gap-1">
			<label for="max-angular-speed" class="text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
				Max Angular Speed (deg/s)
			</label>
			<input
				id="max-angular-speed"
				type="number"
				value={maxAngularSpeed}
				onchange={(e) => setMaxAngularSpeed(Number((e.target as HTMLInputElement).value))}
				min="0"
				max="180"
				disabled={isAutomatic}
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900 disabled:opacity-50"
			/>
		</div>
		<span class="text-[10px] text-slate-400 italic mb-1.5 self-end">Sets joystick sensitivity</span>
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
				disabled={isAutomatic}
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900 disabled:opacity-50"
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
				max={maxAngularSpeed}
				min="0"
				disabled={isAutomatic}
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900 disabled:opacity-50"
			/>
		</div>

		<button
			type="button"
			onclick={spin}
			disabled={spinMutation.isPending || isAutomatic}
			class="rounded bg-slate-800 px-4 py-1.5 text-sm font-medium text-white shadow-sm transition hover:bg-slate-700 disabled:opacity-50 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-white"
		>
			Spin
		</button>

		{#if spinMutation.isError}
			<p class="text-sm text-red-600 dark:text-red-400">{spinMutation.error?.message}</p>
		{/if}
	</div>

	<div class="flex flex-wrap items-end gap-3 rounded-lg border border-slate-200 p-4 dark:border-slate-800">
		<div class="flex flex-col gap-1">
			<label for="distance-mm" class="text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
				Distance (mm)
			</label>
			<input
				id="distance-mm"
				type="number"
				bind:value={distanceMm}
				disabled={isAutomatic}
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900 disabled:opacity-50"
			/>
		</div>

		<div class="flex flex-col gap-1">
			<label for="mm-per-sec" class="text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
				Speed (mm/s)
			</label>
			<input
				id="mm-per-sec"
				type="number"
				bind:value={mmPerSec}
				max={maxLinearSpeed}
				min="0"
				disabled={isAutomatic}
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900 disabled:opacity-50"
			/>
		</div>

		<button
			type="button"
			onclick={moveStraight}
			disabled={moveStraightMutation.isPending || isAutomatic}
			class="rounded bg-slate-800 px-4 py-1.5 text-sm font-medium text-white shadow-sm transition hover:bg-slate-700 disabled:opacity-50 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-white"
		>
			Move Straight
		</button>

		{#if moveStraightMutation.isError}
			<p class="text-sm text-red-600 dark:text-red-400">{moveStraightMutation.error?.message}</p>
		{/if}
	</div>
</div>
