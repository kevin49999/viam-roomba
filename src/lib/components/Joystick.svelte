<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { BASE_CONTEXT_KEY, type BaseContextValue } from '$lib/contexts/base';

	const baseContext = getContext<BaseContextValue>(BASE_CONTEXT_KEY);
	if (!baseContext) {
		throw new Error('Joystick must be used inside BaseProvider');
	}

	const { setVelocityMutation, maxLinearSpeed, maxAngularSpeed, automaticModeQuery } = baseContext;

	const isAutomatic = $derived.by(() => {
		const data = automaticModeQuery.data;
		if (data && typeof data === 'object' && !Array.isArray(data)) {
			return (data as Record<string, unknown>)['automatic_mode'] === true;
		}
		return false;
	});

	let container: HTMLDivElement;
	let knob: SVGCircleElement;
	let isDragging = $state(false);
	let keyboardEnabled = $state(false);
	let keys = $state(new Set<string>());
	let x = $state(0);
	let y = $state(0);
	let size = 150;
	let radius = size / 2;
	let knobRadius = 25;
	let maxDist = radius - knobRadius;

	function handleStart(event: MouseEvent | TouchEvent) {
		if (isAutomatic) return;
		isDragging = true;
		handleMove(event);
	}

	function handleMove(event: MouseEvent | TouchEvent) {
		if (isAutomatic) {
			isDragging = false;
			return;
		}
		if (!isDragging) return;

		const rect = container.getBoundingClientRect();
		const clientX = 'touches' in event ? event.touches[0].clientX : event.clientX;
		const clientY = 'touches' in event ? event.touches[0].clientY : event.clientY;

		let dx = clientX - (rect.left + radius);
		let dy = clientY - (rect.top + radius);

		const dist = Math.sqrt(dx * dx + dy * dy);
		if (dist > maxDist) {
			dx *= maxDist / dist;
			dy *= maxDist / dist;
		}

		x = dx;
		y = dy;

		updateVelocity();
	}

	function handleEnd() {
		isDragging = false;
		x = 0;
		y = 0;
		updateVelocity(true);
	}

	function handleKeyDown(event: KeyboardEvent) {
		if (!keyboardEnabled || isAutomatic) return;
		const key = event.key.toLowerCase();
		if (['w', 'a', 's', 'd'].includes(key)) {
			// Prevent scrolling with WASD if enabled
			event.preventDefault();
			keys.add(key);
			updateKeyboardVelocity();
		}
	}

	function handleKeyUp(event: KeyboardEvent) {
		if (!keyboardEnabled) return;
		const key = event.key.toLowerCase();
		if (['w', 'a', 's', 'd'].includes(key)) {
			keys.delete(key);
			updateKeyboardVelocity();
		}
	}

	function updateKeyboardVelocity() {
		let linearY = 0;
		let angularZ = 0;

		if (keys.has('w')) linearY += 1;
		if (keys.has('s')) linearY -= 1;
		if (keys.has('a')) angularZ += 1;
		if (keys.has('d')) angularZ -= 1;

		x = -angularZ * maxDist;
		y = -linearY * maxDist;

		updateVelocity(keys.size === 0);
	}

	$effect(() => {
		if (!keyboardEnabled || isAutomatic) {
			keys.clear();
			if (x !== 0 || y !== 0) {
				x = 0;
				y = 0;
				updateVelocity(true);
			}
		}
	});

	let lastUpdate = 0;
	const THROTTLE_MS = 100;

	function updateVelocity(force = false) {
		const now = Date.now();
		if (!force && now - lastUpdate < THROTTLE_MS) return;
		lastUpdate = now;

		// Map x/y to velocity
		// x is angular.z (left/right) -> reversed because x+ is right, but usually angular z+ is CCW (left)
		// y is linear.y (forward/backward) -> reversed because y+ is down in screen coords
		
		const linearY = -y / maxDist; // range [-1, 1]
		const angularZ = -x / maxDist; // range [-1, 1]

		setVelocityMutation.mutate([
			{ x: 0, y: Number((linearY * maxLinearSpeed).toFixed(2)), z: 0 },
			{ x: 0, y: 0, z: Number((angularZ * maxAngularSpeed).toFixed(2)) }
		]);
	}

	onMount(() => {
		window.addEventListener('mousemove', handleMove);
		window.addEventListener('mouseup', handleEnd);
		window.addEventListener('touchmove', handleMove, { passive: false });
		window.addEventListener('touchend', handleEnd);
		window.addEventListener('keydown', handleKeyDown);
		window.addEventListener('keyup', handleKeyUp);

		return () => {
			window.removeEventListener('mousemove', handleMove);
			window.removeEventListener('mouseup', handleEnd);
			window.removeEventListener('touchmove', handleMove);
			window.removeEventListener('touchend', handleEnd);
			window.removeEventListener('keydown', handleKeyDown);
			window.removeEventListener('keyup', handleKeyUp);
		};
	});
</script>

<div class="flex flex-col items-center gap-4">
	<div class="flex items-center justify-between w-full">
		<span class="text-xs font-semibold uppercase tracking-wider text-slate-400">Joystick Control</span>
		<label class="flex items-center gap-2 cursor-pointer select-none">
			<input 
				type="checkbox" 
				bind:checked={keyboardEnabled} 
				class="w-3 h-3 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
			/>
			<span class="text-[10px] font-semibold uppercase tracking-wider text-slate-400">Keyboard (WASD)</span>
		</label>
	</div>
	<div
		bind:this={container}
		role="application"
		aria-label="Joystick Control"
		class="relative flex items-center justify-center rounded-full bg-slate-100 dark:bg-slate-800 shadow-inner select-none touch-none {isAutomatic ? 'opacity-50 cursor-not-allowed' : ''}"
		style="width: {size}px; height: {size}px;"
		onmousedown={handleStart}
		ontouchstart={handleStart}
	>
		<svg width={size} height={size} viewBox="0 0 {size} {size}">
			<!-- Outer circle -->
			<circle cx={radius} cy={radius} r={radius - 2} fill="none" stroke="currentColor" stroke-width="2" class="text-slate-200 dark:text-slate-700" />
			
			<!-- Crosshair -->
			<line x1={radius} y1="10" x2={radius} y2={size - 10} stroke="currentColor" stroke-width="1" stroke-dasharray="4" class="text-slate-300 dark:text-slate-600" />
			<line x1="10" y1={radius} x2={size - 10} y2={radius} stroke="currentColor" stroke-width="1" stroke-dasharray="4" class="text-slate-300 dark:text-slate-600" />

			<!-- Knob -->
			<circle
				bind:this={knob}
				cx={radius + x}
				cy={radius + y}
				r={knobRadius}
				class="{isAutomatic ? 'fill-slate-400 dark:fill-slate-600 cursor-not-allowed' : 'fill-blue-600 dark:fill-blue-500 cursor-grab active:cursor-grabbing'} shadow-lg transition-colors duration-200"
			/>
		</svg>
	</div>
	
	<div class="flex gap-4 text-[10px] font-mono text-slate-400">
		<span>Linear Y: {(-y / maxDist).toFixed(2)}</span>
		<span>Angular Z: {(-x / maxDist).toFixed(2)}</span>
	</div>
</div>
