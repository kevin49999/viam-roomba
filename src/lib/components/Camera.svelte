<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { CAMERA_CONTEXT_KEY, type CameraContextValue } from '$lib/contexts/camera';

	const cameraContext = getContext<CameraContextValue>(CAMERA_CONTEXT_KEY);

	let videoElement: HTMLVideoElement | null = $state(null);
	let canvas: HTMLCanvasElement | null = $state(null);
	let stream: MediaStream | null = $state(null);

	$effect(() => {
		const result = cameraContext.query.data;
		if (result && result.images && result.images.length > 0 && canvas) {
			const cameraImage = result.images[0];
			const blob = new Blob([cameraImage.image as any], { type: cameraImage.mimeType });
			const url = URL.createObjectURL(blob);
			const img = new Image();
			img.onload = () => {
				if (canvas) {
					canvas.width = img.width;
					canvas.height = img.height;
					const ctx = canvas.getContext('2d');
					ctx?.drawImage(img, 0, 0);
				}
				URL.revokeObjectURL(url);
			};
			img.src = url;
		}
	});

	onMount(() => {
		if (canvas) {
			// @ts-ignore - captureStream might not be in all types but is standard in browsers
			stream = canvas.captureStream(30);
			if (videoElement) {
				videoElement.srcObject = stream;
			}
		}
	});
</script>

<div class="camera-container relative aspect-video w-full">
	<canvas bind:this={canvas} class="hidden"></canvas>
	<video 
		bind:this={videoElement} 
		autoplay 
		playsinline 
		muted 
		class="h-full w-full object-cover"
	></video>
	
	{#if cameraContext.query.isLoading && !cameraContext.query.data}
		<div class="absolute inset-0 flex items-center justify-center bg-slate-100 dark:bg-slate-800">
			<div class="flex flex-col items-center gap-2">
				<div class="h-6 w-6 animate-spin rounded-full border-2 border-slate-300 border-t-slate-600"></div>
				<span class="text-xs font-medium text-slate-500">Initializing Camera...</span>
			</div>
		</div>
	{:else if !cameraContext.query.isLoading && (!cameraContext.query.data || !cameraContext.query.data.images || cameraContext.query.data.images.length === 0)}
		<div class="absolute inset-0 flex items-center justify-center bg-slate-100 dark:bg-slate-800">
			<span class="text-xs font-medium text-slate-500">No camera feed available</span>
		</div>
	{/if}
</div>

<style>
	.camera-container {
		@apply overflow-hidden rounded-xl border border-slate-200 bg-slate-950 shadow-sm dark:border-slate-800;
	}
</style>
