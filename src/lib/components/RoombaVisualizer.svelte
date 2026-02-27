<script lang="ts">
	import { getContext } from 'svelte';
	import { READINGS_CONTEXT_KEY, type ReadingsContextValue } from '$lib/contexts/readings';

	const readingsContext = getContext<ReadingsContextValue>(READINGS_CONTEXT_KEY);
	if (!readingsContext) {
		throw new Error('RoombaVisualizer must be used inside ReadingsProvider');
	}

	const { query } = readingsContext;

	// Use $derived for reactivity in Svelte 5
	const readings = $derived((query.data as Record<string, any>) || {});
</script>

<div class="flex flex-col items-center h-full">
    <h2 class="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-4 self-start">Sensor Visualization</h2>
    
    <div class="relative w-full aspect-square max-w-[240px] bg-slate-50 dark:bg-slate-900 rounded-2xl flex items-center justify-center p-4 border border-slate-200 dark:border-slate-800">
        <svg viewBox="0 0 200 200" class="w-full h-full">
            <!-- Wall Sensor -->
            <rect 
                x="50" y="5" width="100" height="10" rx="2"
                fill={readings.wall ? '#EAB308' : '#cbd5e1'} 
                class="transition-colors duration-200"
            />
            <text x="100" y="25" text-anchor="middle" font-size="8" class="fill-slate-400 font-semibold uppercase tracking-wider">Wall</text>

            <!-- Cliff Sensors -->
            <!-- Cliff Left -->
            <rect 
                x="10" y="85" width="10" height="30" rx="2"
                fill={readings.cliff_left ? '#EAB308' : '#cbd5e1'} 
                class="transition-colors duration-200"
            />
            <!-- Cliff Front Left -->
            <rect 
                x="40" y="35" width="30" height="10" rx="2"
                fill={readings.cliff_front_left ? '#EAB308' : '#cbd5e1'} 
                class="transition-colors duration-200"
            />
            <!-- Cliff Front Right -->
            <rect 
                x="130" y="35" width="30" height="10" rx="2"
                fill={readings.cliff_front_right ? '#EAB308' : '#cbd5e1'} 
                class="transition-colors duration-200"
            />
            <!-- Cliff Right -->
            <rect 
                x="180" y="85" width="10" height="30" rx="2"
                fill={readings.cliff_right ? '#EAB308' : '#cbd5e1'} 
                class="transition-colors duration-200"
            />

            <!-- Roomba Circle -->
            <circle cx="100" cy="110" r="60" class="fill-slate-800 dark:fill-slate-700" />
            
            <!-- Bump Left Arc -->
            <path 
                d="M 40 110 A 60 60 0 0 1 100 50 L 100 110 Z" 
                fill={readings.bump_left ? '#EF4444' : 'transparent'} 
                class="transition-colors duration-200"
            />
            
            <!-- Bump Right Arc -->
            <path 
                d="M 100 50 A 60 60 0 0 1 160 110 L 100 110 Z" 
                fill={readings.bump_right ? '#EF4444' : 'transparent'} 
                class="transition-colors duration-200"
            />
            
            <!-- Inner design for Roomba -->
            <circle cx="100" cy="110" r="55" class="fill-slate-700 dark:fill-slate-600" />
            <circle cx="100" cy="110" r="20" class="fill-slate-600 dark:fill-slate-500" />
            <text x="100" y="114" text-anchor="middle" fill="white" font-size="8" class="font-bold">ROOMBA</text>
        </svg>
    </div>

    <div class="mt-6 grid grid-cols-2 gap-4 w-full text-xs font-mono uppercase tracking-tight">
        <div class="flex items-center gap-2">
            <div class="w-3 h-3 rounded-full {readings.bump_left ? 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.6)]' : 'bg-slate-200 dark:bg-slate-700'} transition-all duration-300"></div>
            <span class="{readings.bump_left ? 'text-red-500 font-bold' : 'text-slate-400'}">Bump Left</span>
        </div>
        <div class="flex items-center gap-2">
            <div class="w-3 h-3 rounded-full {readings.bump_right ? 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.6)]' : 'bg-slate-200 dark:bg-slate-700'} transition-all duration-300"></div>
            <span class="{readings.bump_right ? 'text-red-500 font-bold' : 'text-slate-400'}">Bump Right</span>
        </div>
        <div class="flex items-center gap-2">
            <div class="w-3 h-3 rounded-full {readings.wall ? 'bg-yellow-500 shadow-[0_0_8px_rgba(234,179,8,0.6)]' : 'bg-slate-200 dark:bg-slate-700'} transition-all duration-300"></div>
            <span class="{readings.wall ? 'text-yellow-600 font-bold' : 'text-slate-400'}">Wall</span>
        </div>
        <div class="flex items-center gap-2">
            <div class="w-3 h-3 rounded-full {(readings.cliff_left || readings.cliff_front_left || readings.cliff_front_right || readings.cliff_right) ? 'bg-yellow-500 shadow-[0_0_8px_rgba(234,179,8,0.6)]' : 'bg-slate-200 dark:bg-slate-700'} transition-all duration-300"></div>
            <span class="{(readings.cliff_left || readings.cliff_front_left || readings.cliff_front_right || readings.cliff_right) ? 'text-yellow-600 font-bold' : 'text-slate-400'}">Cliff</span>
        </div>
    </div>
</div>
