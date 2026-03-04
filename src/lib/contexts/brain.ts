import type { CreateMutationResult, QueryObserverResult } from '@tanstack/svelte-query';
import type { JsonValue } from '@viamrobotics/sdk';

export type BrainCommandPayload = Record<string, JsonValue>;

export interface BrainContextValue {
	/** Mutation to send doCommand to the brain service. Call with [command]. */
	doCommandMutation: CreateMutationResult<JsonValue, Error, [command: BrainCommandPayload], unknown>;
	/** Query for current automatic mode state (polled every 5s). */
	automaticModeQuery: QueryObserverResult<JsonValue, Error>;
	/** Toggle automatic mode on/off. */
	toggleAutomaticMode: () => void;
}

export const BRAIN_CONTEXT_KEY = Symbol('brain');
