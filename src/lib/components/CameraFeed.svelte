<script lang="ts">
	import { CameraStream, useConnectionStatus, useResourceNames } from '@viamrobotics/svelte-sdk';
	import { DEFAULT_PART_ID } from '$lib/consts';
	import { MachineConnectionEvent } from '@viamrobotics/sdk';

	interface Props {
		name: string;
		partID: string;
	}

	let { name, partID }: Props = $props();

	const connectionStatus = useConnectionStatus(() => DEFAULT_PART_ID);
	const resourceNames = useResourceNames(() => DEFAULT_PART_ID);
</script>
{#if connectionStatus.current === MachineConnectionEvent.CONNECTED && resourceNames.current.map(resource => resource.name).includes(name)}
	<CameraStream {name} {partID} />
{:else}
	<div class="flex items-center justify-center h-full">
		<div class="text-sm text-slate-500">Connecting to camera...</div>
	</div>
{/if}
