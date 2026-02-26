<script lang="ts">
	import { createResourceClient, createResourceQuery } from '@viamrobotics/svelte-sdk';
	import { CameraClient } from '@viamrobotics/sdk';
	import { setContext } from 'svelte';
	import { CAMERA_CONTEXT_KEY, type CameraContextValue } from '$lib/contexts/camera';

	interface Props {
		partID: string;
		name: string;
		refetchInterval?: number;
		children?: import('svelte').Snippet;
	}
	const FPS_30 = 1000 / 30;

	let { partID, name, refetchInterval = FPS_30, children }: Props = $props();

	const cameraClient = createResourceClient(CameraClient, () => partID, () => name);
	const query = createResourceQuery(cameraClient, 'getImages', () => ({
		refetchInterval
	}));

	const contextValue: CameraContextValue = {
		query,
		refetch: () => {
			query.refetch();
		}
	};
	setContext(CAMERA_CONTEXT_KEY, contextValue);
</script>

{#if children}
	{@render children()}
{/if}
