import type { CreateQueryResult } from '@tanstack/svelte-query';

export interface CameraImage {
	sourceName: string;
	image: Uint8Array;
	mimeType: string;
}

export interface GetImagesResponse {
	images: CameraImage[];
	metadata: any;
}

export interface CameraContextValue {
	/** Reactive query result for getImages() */
	query: CreateQueryResult<GetImagesResponse>;
	/** Trigger a refetch of the images */
	refetch: () => void;
}

export const CAMERA_CONTEXT_KEY = Symbol('camera');
