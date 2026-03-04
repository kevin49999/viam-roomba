<script lang="ts">
	import { createResourceClient, createResourceMutation } from '@viamrobotics/svelte-sdk';
	import { BaseClient } from '@viamrobotics/sdk';
	import { setContext } from 'svelte';
	import { BASE_CONTEXT_KEY, type BaseContextValue } from '$lib/contexts/base';

	interface Props {
		partID: string;
		name: string;
		children?: import('svelte').Snippet;
	}

	let { partID, name, children }: Props = $props();

	const baseClient = createResourceClient(BaseClient, () => partID, () => name);
	const spinMutation = createResourceMutation(baseClient, 'spin');
	const moveStraightMutation = createResourceMutation(baseClient, 'moveStraight');
	const setVelocityMutation = createResourceMutation(baseClient, 'setVelocity');

	let maxLinearSpeed = $state(300);
	let maxAngularSpeed = $state(90);

	const contextValue: BaseContextValue = {
		spinMutation: spinMutation as unknown as BaseContextValue['spinMutation'],
		moveStraightMutation: moveStraightMutation as unknown as BaseContextValue['moveStraightMutation'],
		setVelocityMutation: setVelocityMutation as unknown as BaseContextValue['setVelocityMutation'],
		get maxLinearSpeed() {
			return maxLinearSpeed;
		},
		get maxAngularSpeed() {
			return maxAngularSpeed;
		},
		setMaxLinearSpeed: (speed: number) => {
			maxLinearSpeed = Math.min(Math.max(0, speed), 500); // 500 mm/s is a reasonable absolute max
		},
		setMaxAngularSpeed: (speed: number) => {
			maxAngularSpeed = Math.min(Math.max(0, speed), 180); // 180 deg/s is a reasonable absolute max
		}
	};
	setContext(BASE_CONTEXT_KEY, contextValue);
</script>

{#if children}
	{@render children()}
{/if}
