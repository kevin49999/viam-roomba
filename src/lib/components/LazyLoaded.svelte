<script lang="ts">
	import { browser } from '$app/environment';
	import type { Snippet } from 'svelte';

	interface Props {
		componentImportFunction: () => Promise<any>;
		loadingSnippet?: Snippet;
		errorSnippet?: Snippet<[error: Error]>;
		[key: string]: any;
	}
	let { componentImportFunction, loadingSnippet, errorSnippet, ...rest }: Props = $props();

	let componentPromise = $state.raw<Promise<any> | null>(null);

	$effect(() => {
		if (browser) {
			componentPromise = componentImportFunction();
		}
	});
</script>

{#if componentPromise}
	{#await componentPromise}
		{#if loadingSnippet}
			{@render loadingSnippet()}
		{:else}
			<p>Loading...</p>
		{/if}
	{:then { default: Component }}
		<Component {...rest} />
	{:catch error}
		{#if errorSnippet}
			{@render errorSnippet(error)}
		{:else}
			<p class="text-red-600">Error loading component: {error.message}</p>
		{/if}
	{/await}
{/if}