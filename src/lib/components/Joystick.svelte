<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { BASE_CONTEXT_KEY, type BaseContextValue } from '$lib/contexts/base';

	const baseContext = getContext<BaseContextValue>(BASE_CONTEXT_KEY);
	if (!baseContext) {
		throw new Error('Joystick must be used inside BaseProvider');
	}

	const { setVelocityMutation } = baseContext;

	let container: HTMLDivElement;
	let knob: SVGCircleElement;
	let isDragging = $state(false);
	let x = $state(0);
	let y = $state(0);
	let size = 150;
	let radius = size / 2;
	let knobRadius = 25;
	let maxDist = radius - knobRadius;

	function handleStart(event: MouseEvent | TouchEvent) {
		isDragging = true;
		handleMove(event);
	}

	function handleMove(event: MouseEvent | TouchEvent) {
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

		// You might want to scale these by some max speed
		const maxLinearSpeed = 300; // mm/s
		const maxAngularSpeed = 90; // deg/s

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

		return () => {
			window.removeEventListener('mousemove', handleMove);
			window.removeEventListener('mouseup', handleEnd);
			window.removeEventListener('touchmove', handleMove);
			window.removeEventListener('touchend', handleEnd);
		};
	});
</script>

<div class="flex flex-col items-center gap-4">
	<span class="text-xs font-semibold uppercase tracking-wider text-slate-400">Joystick Control</span>
	<div
		bind:this={container}
		role="application"
		aria-label="Joystick Control"
		class="relative flex items-center justify-center rounded-full bg-slate-100 dark:bg-slate-800 shadow-inner select-none touch-none"
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
				class="fill-blue-600 dark:fill-blue-500 shadow-lg cursor-grab active:cursor-grabbing transition-colors duration-200"
			/>
		</svg>
	</div>
	
	<div class="flex gap-4 text-[10px] font-mono text-slate-400">
		<span>Linear Y: {(-y / maxDist).toFixed(2)}</span>
		<span>Angular Z: {(-x / maxDist).toFixed(2)}</span>
	</div>
</div>
