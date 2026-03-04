import type { CreateMutationResult } from '@tanstack/svelte-query';
import type { Vector3 } from '@viamrobotics/sdk';


export interface BaseContextValue {
	/** Mutation to send spin to the base. Call with [angleDeg, degsPerSec] or [angleDeg, degsPerSec, extra, callOptions] */
	spinMutation: CreateMutationResult<void, Error, [angleDeg: number, degsPerSec: number], unknown>;
	/** Mutation to send moveStraight to the base. Call with [distanceMm, mmPerSec] or [distanceMm, mmPerSec, extra, callOptions] */
	moveStraightMutation: CreateMutationResult<void, Error, [distanceMm: number, mmPerSec: number], unknown>;
	/** Mutation to send setVelocity to the base. Call with [linear, angular] or [linear, angular, extra, callOptions] */
	setVelocityMutation: CreateMutationResult<void, Error, [linear: Vector3, angular: Vector3], unknown>;
	/** Maximum linear speed in mm/s */
	maxLinearSpeed: number;
	/** Maximum angular speed in deg/s */
	maxAngularSpeed: number;
	/** Setter for maxLinearSpeed */
	setMaxLinearSpeed: (speed: number) => void;
	/** Setter for maxAngularSpeed */
	setMaxAngularSpeed: (speed: number) => void;
}


export const BASE_CONTEXT_KEY = Symbol('base');