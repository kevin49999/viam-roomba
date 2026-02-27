<script lang="ts">
	import { getContext } from 'svelte';
	import { READINGS_CONTEXT_KEY, type ReadingsContextValue } from '$lib/contexts/readings';

	const readingsContext = getContext<ReadingsContextValue>(READINGS_CONTEXT_KEY);
	if (!readingsContext) {
		throw new Error('UltraSonic must be used inside ReadingsProvider');
	}

	const { query } = readingsContext;

	/** dist from readings in meters; undefined if missing or invalid */
	const distance = $derived.by(() => {
		const data = query.data;
		if (data == null || typeof data !== 'object' || Array.isArray(data)) return undefined;
		const raw = (data as Record<string, unknown>)['distance'];
		if (typeof raw !== 'number') return undefined;
		return raw;
	});
</script>

<div class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 shadow-sm">
	<div class="flex items-center justify-center w-5 h-5 rounded-full bg-blue-500 text-white">
		<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
			<path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/>
		</svg>
	</div>
	<div class="flex flex-col">
		<span class="text-[10px] font-bold uppercase tracking-wider text-slate-400 leading-none">Ultrasonic</span>
		<span class="text-sm font-mono font-bold text-slate-700 dark:text-slate-200 leading-tight">
			{distance != null ? `${distance.toFixed(3)}m` : '—'}
		</span>
	</div>
</div>
