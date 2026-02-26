import type { CreateMutationResult } from '@tanstack/svelte-query';
import type { JsonValue } from '@viamrobotics/sdk';

/**
 * Command payload for doCommand. The base (e.g. Roomba) expects a "command" key
 * with a string value (e.g. "seek_dock", "clean", "enter_safe_mode").
 */
export type DoCommandPayload = Record<string, JsonValue>;

export interface BaseContextValue {
	/** Mutation to send doCommand to the base. Call with .mutate([command]) or .mutate([command, callOptions]) */
	doCommandMutation: CreateMutationResult<JsonValue, Error, [command: DoCommandPayload], unknown>;
}

export const BASE_CONTEXT_KEY = Symbol('base');