<script lang="ts">
	import { getContext } from 'svelte';
	import { BASE_CONTEXT_KEY, type BaseContextValue } from '$lib/contexts/base';

	const baseContext = getContext<BaseContextValue>(BASE_CONTEXT_KEY);
	if (!baseContext) {
		throw new Error('BaseCommands must be used inside BaseProvider');
	}

	const { spinMutation, moveStraightMutation, maxLinearSpeed, maxAngularSpeed, setMaxLinearSpeed, setMaxAngularSpeed } = baseContext;

	let angleDeg = $state(90);
	let degsPerSec = $state(45);

	let distanceMm = $state(500);
	let mmPerSec = $state(100);

	function spin() {
		spinMutation.mutate([angleDeg, degsPerSec]);
	}

	function moveStraight() {
		moveStraightMutation.mutate([distanceMm, mmPerSec]);
	}
</script>

<div class="flex flex-col gap-4">
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
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900 disabled:opacity-50"
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

	<div class="flex flex-wrap items-end gap-3 rounded-lg border border-slate-200 p-4 dark:border-slate-800">
		<div class="flex flex-col gap-1">
			<label for="distance-mm" class="text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
				Distance (mm)
			</label>
			<input
				id="distance-mm"
				type="number"
				bind:value={distanceMm}
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
				class="w-24 rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900 disabled:opacity-50"
			/>
		</div>

		<button
			type="button"
			onclick={moveStraight}
			disabled={moveStraightMutation.isPending}
			class="rounded bg-slate-800 px-4 py-1.5 text-sm font-medium text-white shadow-sm transition hover:bg-slate-700 disabled:opacity-50 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-white"
		>
			Move Straight
		</button>

		{#if moveStraightMutation.isError}
			<p class="text-sm text-red-600 dark:text-red-400">{moveStraightMutation.error?.message}</p>
		{/if}
	</div>
</div>
