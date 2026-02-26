import type { CreateMutationResult } from '@tanstack/svelte-query';
import type { JsonValue, Vector3 } from '@viamrobotics/sdk';

/**
 * Command payload for doCommand. The base (e.g. Roomba) expects a "command" key
 * with a string value (e.g. "seek_dock", "clean", "enter_safe_mode").
 */
export type DoCommandPayload = Record<string, JsonValue>;

export interface BaseContextValue {
	/** Mutation to send doCommand to the base. Call with [command] or [command, callOptions] */
	doCommandMutation: CreateMutationResult<JsonValue, Error, [command: DoCommandPayload], unknown>;
	/** Mutation to send spin to the base. Call with [angleDeg, degsPerSec] or [angleDeg, degsPerSec, extra, callOptions] */
	spinMutation: CreateMutationResult<void, Error, [angleDeg: number, degsPerSec: number], unknown>;
	/** Mutation to send moveStraight to the base. Call with [distanceMm, mmPerSec] or [distanceMm, mmPerSec, extra, callOptions] */
	moveStraightMutation: CreateMutationResult<void, Error, [distanceMm: number, mmPerSec: number], unknown>;
	/** Mutation to send setVelocity to the base. Call with [linear, angular] or [linear, angular, extra, callOptions] */
	setVelocityMutation: CreateMutationResult<void, Error, [linear: Vector3, angular: Vector3], unknown>;
}

export const BASE_CONTEXT_KEY = Symbol('base');