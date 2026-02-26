<script lang="ts">
	import { DEFAULT_PART_ID } from '$lib/consts';
	import type { DialConf } from '@viamrobotics/sdk';
	import { ViamProvider } from '@viamrobotics/svelte-sdk';
	import { getCookie } from 'typescript-cookie';
	import AllReadings from './AllReadings.svelte';
	import BaseCommands from './BaseCommands.svelte';
	import BaseProvider from './BaseProvider.svelte';
	import BatteryPercent from './BatteryPercent.svelte';
	import OIMode from './OIMode.svelte';
	import ReadingsProvider from './ReadingsProvider.svelte';

	let { host, credentials, machineId } = $props<{
		host: string;
		credentials: any;
		machineId?: string | null;
	}>();

	const dialConfigs: Record<string, DialConf> = $derived({
		[DEFAULT_PART_ID]: {
			host: host,
			credentials: credentials,
			signalingAddress: getCookie('is_local')
				? `http://localhost:8081` // Hard-coded to the custom viam-server bind port
				: 'https://app.viam.com',
			disableSessions: false
		}
	});

	$effect(() => {
		console.log(
			`Connecting to ${host} with credentials: ${JSON.stringify(credentials, undefined, 4)}`
		);
	});
</script>

<ViamProvider {dialConfigs}>
	<ReadingsProvider partID={DEFAULT_PART_ID} name="sensor">
		<BaseProvider partID={DEFAULT_PART_ID} name="base">
			<div class="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900/50">
				<div class="flex flex-wrap items-center justify-between gap-4">
					<div>
						<h1 class="text-xl font-bold text-slate-900 dark:text-white">Roomba Dashboard</h1>
						<p class="text-xs text-slate-500 font-mono mt-1">{host}</p>
					</div>
					<BatteryPercent />
				</div>

				<div class="h-px bg-slate-100 dark:bg-slate-800"></div>

				<div class="flex flex-wrap items-center gap-6">
					<div class="flex flex-col gap-2">
						<span class="text-xs font-semibold uppercase tracking-wider text-slate-400">Control Mode</span>
						<OIMode />
					</div>
					
					<div class="flex flex-col gap-2">
						<span class="text-xs font-semibold uppercase tracking-wider text-slate-400">Actions</span>
						<BaseCommands />
					</div>
				</div>
			</div>
			
			<div class="mt-6">
				<AllReadings />
			</div>
		</BaseProvider>
	</ReadingsProvider>
</ViamProvider>
