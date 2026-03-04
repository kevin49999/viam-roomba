<script lang="ts">
	import { getContext } from 'svelte';
	import { BRAIN_CONTEXT_KEY, type BrainContextValue } from '$lib/contexts/brain';

	const brainContext = getContext<BrainContextValue>(BRAIN_CONTEXT_KEY);
	if (!brainContext) {
		throw new Error('AutomaticMode must be used inside BrainProvider');
	}

	const { automaticModeQuery, toggleAutomaticMode, doCommandMutation } = brainContext;

	const isAutomatic = $derived.by(() => {
		const data = automaticModeQuery.data;
		if (data && typeof data === 'object' && !Array.isArray(data)) {
			return (data as Record<string, unknown>)['automatic_mode'] === true;
		}
		return false;
	});
</script>

<div class="flex items-center gap-3">
	<button
		type="button"
		onclick={toggleAutomaticMode}
		disabled={doCommandMutation.isPending}
		class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-blue-600 focus:ring-offset-2 {isAutomatic ? 'bg-blue-600' : 'bg-slate-200 dark:bg-slate-700'}"
		role="switch"
		aria-checked={isAutomatic}
	>
		<span
			aria-hidden="true"
			class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out {isAutomatic ? 'translate-x-5' : 'translate-x-0'}"
		></span>
	</button>
	<span class="text-sm font-medium text-slate-700 dark:text-slate-300">
		Automatic Mode {isAutomatic ? 'On' : 'Off'}
	</span>
</div>
