<script lang="ts">
	import { onMount, onDestroy, getContext } from 'svelte';
	import { BRAIN_CONTEXT_KEY, type BrainContextValue } from '$lib/contexts/brain';

	interface Note {
		id: number;
		name: string;
		frequency: number;
	}

	interface SongNote extends Note {
		duration: number; // Roomba duration byte: 8=1/8, 16=1/4, 32=1/2, 64=whole
	}

	interface KeyLayout extends Note {
		isBlack: boolean;
		x: number;
	}

	const DURATION_OPTIONS = [
		{ label: '1/8', value: 8, tone: '8n', secs: 0.3 },
		{ label: '1/4', value: 16, tone: '4n', secs: 0.55 },
		{ label: '1/2', value: 32, tone: '2n', secs: 1.05 },
		{ label: 'W', value: 64, tone: '1n', secs: 2.05 }
	] as const;

	type DurationValue = (typeof DURATION_OPTIONS)[number]['value'];

	function getDurationOption(value: number) {
		return DURATION_OPTIONS.find((d) => d.value === value) ?? DURATION_OPTIONS[1];
	}

	const ALL_NOTES: Note[] = [
		{ id: 31, name: 'G', frequency: 49.0 },
		{ id: 32, name: 'G#', frequency: 51.9 },
		{ id: 33, name: 'A', frequency: 55.0 },
		{ id: 34, name: 'A#', frequency: 58.3 },
		{ id: 35, name: 'B', frequency: 61.7 },
		{ id: 36, name: 'C', frequency: 65.4 },
		{ id: 37, name: 'C#', frequency: 69.3 },
		{ id: 38, name: 'D', frequency: 73.4 },
		{ id: 39, name: 'D#', frequency: 77.8 },
		{ id: 40, name: 'E', frequency: 82.4 },
		{ id: 41, name: 'F', frequency: 87.3 },
		{ id: 42, name: 'F#', frequency: 92.5 },
		{ id: 43, name: 'G', frequency: 98.0 },
		{ id: 44, name: 'G#', frequency: 103.8 },
		{ id: 45, name: 'A', frequency: 110.0 },
		{ id: 46, name: 'A#', frequency: 116.5 },
		{ id: 47, name: 'B', frequency: 123.5 },
		{ id: 48, name: 'C', frequency: 130.8 },
		{ id: 49, name: 'C#', frequency: 138.6 },
		{ id: 50, name: 'D', frequency: 146.8 },
		{ id: 51, name: 'D#', frequency: 155.6 },
		{ id: 52, name: 'E', frequency: 164.8 },
		{ id: 53, name: 'F', frequency: 174.6 },
		{ id: 54, name: 'F#', frequency: 185.0 },
		{ id: 55, name: 'G', frequency: 196.0 },
		{ id: 56, name: 'G#', frequency: 207.7 },
		{ id: 57, name: 'A', frequency: 220.0 },
		{ id: 58, name: 'A#', frequency: 233.1 },
		{ id: 59, name: 'B', frequency: 246.9 },
		{ id: 60, name: 'C', frequency: 261.6 },
		{ id: 61, name: 'C#', frequency: 277.2 },
		{ id: 62, name: 'D', frequency: 293.7 },
		{ id: 63, name: 'D#', frequency: 311.1 },
		{ id: 64, name: 'E', frequency: 329.6 },
		{ id: 65, name: 'F', frequency: 349.2 },
		{ id: 66, name: 'F#', frequency: 370.0 },
		{ id: 67, name: 'G', frequency: 392.0 },
		{ id: 68, name: 'G#', frequency: 415.3 },
		{ id: 69, name: 'A', frequency: 440.0 },
		{ id: 70, name: 'A#', frequency: 466.2 },
		{ id: 71, name: 'B', frequency: 493.9 },
		{ id: 72, name: 'C', frequency: 523.3 },
		{ id: 73, name: 'C#', frequency: 554.4 },
		{ id: 74, name: 'D', frequency: 587.3 },
		{ id: 75, name: 'D#', frequency: 622.3 },
		{ id: 76, name: 'E', frequency: 659.3 },
		{ id: 77, name: 'F', frequency: 698.5 },
		{ id: 78, name: 'F#', frequency: 740.0 },
		{ id: 79, name: 'G', frequency: 784.0 },
		{ id: 80, name: 'G#', frequency: 830.6 },
		{ id: 81, name: 'A', frequency: 880.0 },
		{ id: 82, name: 'A#', frequency: 932.4 },
		{ id: 83, name: 'B', frequency: 987.8 },
		{ id: 84, name: 'C', frequency: 1046.5 },
		{ id: 85, name: 'C#', frequency: 1108.8 },
		{ id: 86, name: 'D', frequency: 1174.7 },
		{ id: 87, name: 'D#', frequency: 1244.5 },
		{ id: 88, name: 'E', frequency: 1318.5 },
		{ id: 89, name: 'F', frequency: 1396.9 },
		{ id: 90, name: 'F#', frequency: 1480.0 },
		{ id: 91, name: 'G', frequency: 1568.0 },
		{ id: 92, name: 'G#', frequency: 1661.3 },
		{ id: 93, name: 'A', frequency: 1760.0 },
		{ id: 94, name: 'A#', frequency: 1864.7 },
		{ id: 95, name: 'B', frequency: 1975.6 },
		{ id: 96, name: 'C', frequency: 2093.1 },
		{ id: 97, name: 'C#', frequency: 2217.5 },
		{ id: 98, name: 'D', frequency: 2349.4 },
		{ id: 99, name: 'D#', frequency: 2489.1 },
		{ id: 100, name: 'E', frequency: 2637.1 },
		{ id: 101, name: 'F', frequency: 2793.9 },
		{ id: 102, name: 'F#', frequency: 2960.0 },
		{ id: 103, name: 'G', frequency: 3136.0 },
		{ id: 104, name: 'G#', frequency: 3322.5 },
		{ id: 105, name: 'A', frequency: 3520.1 },
		{ id: 106, name: 'A#', frequency: 3729.4 },
		{ id: 107, name: 'B', frequency: 3951.2 }
	];

	// Piano key dimensions
	const WHITE_KEY_WIDTH = 40;
	const WHITE_KEY_HEIGHT = 130;
	const BLACK_KEY_WIDTH = 25;
	const BLACK_KEY_HEIGHT = 82;

	function buildLayout(notes: Note[]): KeyLayout[] {
		let wi = 0;
		return notes.map((note) => {
			const isBlack = note.name.includes('#');
			let x: number;
			if (isBlack) {
				x = wi * WHITE_KEY_WIDTH - BLACK_KEY_WIDTH / 2;
			} else {
				x = wi * WHITE_KEY_WIDTH;
				wi++;
			}
			return { ...note, isBlack, x };
		});
	}

	const layout = buildLayout(ALL_NOTES);
	const whiteKeys = layout.filter((k) => !k.isBlack);
	const blackKeys = layout.filter((k) => k.isBlack);
	const totalWidth = whiteKeys.length * WHITE_KEY_WIDTH;

	// MIDI octave: C4 = note 60 → octave = floor(id/12) - 1
	function getOctave(id: number): number {
		return Math.floor(id / 12) - 1;
	}

	const brainContext = getContext<BrainContextValue>(BRAIN_CONTEXT_KEY);
	const { doCommandMutation } = brainContext ?? {};

	// Song and playback state
	let song: SongNote[] = $state([]);
	let recording = $state(false);
	let playing = $state(false);
	let activeId = $state<number | null>(null);
	let playSlot = $state<number | null>(null);
	let songSlot = $state(0); // 0-4
	let selectedDuration = $state<DurationValue>(16); // default 1/4 note

	function sendSong() {
		if (!song.length || !doCommandMutation) return;
		const songBytes: number[] = [songSlot, song.length];
		for (const note of song) {
			songBytes.push(note.id, note.duration);
		}
		doCommandMutation.mutate([{ command: 'add_song', song_bytes: songBytes }]);
	}

	// Tone.js — lazy loaded to avoid SSR issues
	let synth: any = null;
	let ToneLib: any = null;

	async function initTone() {
		if (ToneLib) return;
		ToneLib = await import('tone');
		await ToneLib.start();
		synth = new ToneLib.PolySynth(ToneLib.Synth, {
			oscillator: { type: 'triangle' },
			envelope: { attack: 0.01, decay: 0.2, sustain: 0.3, release: 1.0 }
		}).toDestination();
	}

	async function playNote(note: Note) {
		activeId = note.id;
		const id = note.id;
		const dur = getDurationOption(selectedDuration);
		setTimeout(() => {
			if (activeId === id) activeId = null;
		}, dur.secs * 1000);

		try {
			await initTone();
			synth.triggerAttackRelease(note.frequency, dur.tone);
		} catch (e) {
			console.error('playNote error:', e);
		}

		if (recording && song.length < 16) {
			song = [...song, { ...note, duration: selectedDuration }];
		}
	}

	async function playSong() {
		if (!song.length || playing) return;
		playing = true;
		playSlot = null;

		try {
			await initTone();
			const now = ToneLib.now();
			const snapshot = [...song];

			let offset = 0;
			snapshot.forEach((note, i) => {
				const dur = getDurationOption(note.duration);
				synth.triggerAttackRelease(note.frequency, dur.tone, now + offset);
				const noteOffset = offset;
				setTimeout(() => {
					playSlot = i;
				}, noteOffset * 1000);
				offset += dur.secs;
			});

			setTimeout(
				() => {
					playing = false;
					playSlot = null;
				},
				offset * 1000 + 600
			);
		} catch (e) {
			console.error('playSong error:', e);
			playing = false;
			playSlot = null;
		}
	}

	let keyboardElement: HTMLElement | undefined;

	onMount(() => {
		// Default view: center on C4 (note 60)
		if (keyboardElement) {
			const c4 = layout.find((k) => k.id === 60);
			if (c4) keyboardElement.scrollLeft = Math.max(0, c4.x - keyboardElement.clientWidth / 2 + WHITE_KEY_WIDTH);
		}
	});

	onDestroy(() => {
		try {
			synth?.dispose();
		} catch {}
	});
</script>

<div
	class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900/50"
>
	<!-- Header -->
	<div class="mb-4 flex items-center justify-between flex-wrap gap-2">
		<h2 class="text-sm font-semibold uppercase tracking-wider text-slate-400">MIDI Keyboard</h2>
		<div class="flex items-center gap-2 flex-wrap">
			<!-- Duration selector -->
			<div class="flex items-center gap-1">
				<span class="text-xs text-slate-400">Duration:</span>
				{#each DURATION_OPTIONS as opt}
					<button
						onclick={() => (selectedDuration = opt.value)}
						class="rounded px-2 py-1 text-xs font-medium transition-colors {selectedDuration === opt.value
							? 'bg-indigo-500 text-white'
							: 'bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700'}"
					>
						{opt.label}
					</button>
				{/each}
			</div>
			<button
				onclick={() => (recording = !recording)}
				class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium transition-colors {recording
					? 'bg-red-500 text-white hover:bg-red-600'
					: 'bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700'}"
			>
				<span
					class="h-2 w-2 rounded-full {recording ? 'animate-pulse bg-white' : 'bg-slate-400'}"
				></span>
				{recording ? 'Recording' : 'Record'}
			</button>
		</div>
	</div>

	<!-- Song builder -->
	<div class="mb-4">
		<div class="mb-2 flex items-center justify-between">
			<span class="text-xs font-medium text-slate-500 dark:text-slate-400"
				>Song ({song.length}/16 notes)</span
			>
			<div class="flex gap-2 flex-wrap">
				<button
					onclick={playSong}
					disabled={!song.length || playing}
					class="rounded-lg px-3 py-1.5 text-xs font-medium bg-indigo-500 text-white hover:bg-indigo-600 disabled:cursor-not-allowed disabled:opacity-40 transition-colors"
				>
					{playing ? 'Playing…' : '▶ Play Song'}
				</button>
				<button
					onclick={() => {
						song = [];
					}}
					disabled={!song.length}
					class="rounded-lg px-3 py-1.5 text-xs font-medium bg-slate-100 text-slate-600 hover:bg-slate-200 disabled:cursor-not-allowed disabled:opacity-40 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700 transition-colors"
				>
					Clear
				</button>
				{#if doCommandMutation}
					<div class="flex items-center gap-1">
						<select
							bind:value={songSlot}
							class="rounded-lg border border-slate-200 bg-white px-2 py-1.5 text-xs text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-300"
						>
							{#each [0, 1, 2, 3, 4] as n}
								<option value={n}>Slot {n}</option>
							{/each}
						</select>
						<button
							onclick={sendSong}
							disabled={!song.length || doCommandMutation.isPending}
							class="rounded-lg px-3 py-1.5 text-xs font-medium bg-emerald-500 text-white hover:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-40 transition-colors"
						>
							Add Song to Roomba
						</button>
					</div>
				{/if}
			</div>
		</div>

		<!-- 16 note slots -->
		<div class="flex min-h-11 flex-wrap gap-1.5">
			{#each { length: 16 } as _, i}
				{#if i < song.length}
					<button
						onclick={() => {
							song = song.filter((_, j) => j !== i);
						}}
						title="Click to remove note {i + 1}"
						class="flex h-11 w-11 flex-col items-center justify-center rounded-lg border text-xs font-bold transition-all {playSlot ===
						i
							? 'scale-110 border-indigo-400 bg-indigo-500 text-white shadow-md'
							: 'border-slate-200 bg-slate-50 text-slate-700 hover:border-red-300 hover:bg-red-50 hover:text-red-600 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-red-900/30 dark:hover:text-red-400'}"
					>
						<span class="leading-none">{song[i].name}</span>
						<span class="text-[9px] font-normal opacity-70">{getDurationOption(song[i].duration).label}</span>
					</button>
				{:else}
					<div
						class="flex h-11 w-11 items-center justify-center rounded-lg border border-dashed border-slate-200 dark:border-slate-700"
					>
						<span class="text-[10px] text-slate-300 dark:text-slate-600">{i + 1}</span>
					</div>
				{/if}
			{/each}
		</div>

		<!-- Song note IDs (for reference) -->
		{#if song.length > 0}
			<p class="mt-2 text-[10px] text-slate-400 font-mono">
				IDs: [{song.map((n) => n.id).join(', ')}]
			</p>
		{/if}
	</div>

	<!-- Piano keyboard -->
	<div bind:this={keyboardElement} class="overflow-x-auto rounded-lg bg-slate-100 p-3 dark:bg-slate-800">
		<div class="relative" style="width: {totalWidth}px; height: {WHITE_KEY_HEIGHT + 4}px;">
			<!-- White keys (rendered first, behind black keys) -->
			{#each whiteKeys as key}
				<button
					onclick={() => playNote(key)}
					class="absolute bottom-0 select-none rounded-b-md border transition-colors {activeId ===
					key.id
						? 'border-indigo-300 bg-indigo-100 dark:bg-indigo-900/60'
						: 'border-slate-300 bg-white hover:bg-slate-50 dark:border-slate-500 dark:bg-slate-50 dark:hover:bg-slate-200'}"
					style="left: {key.x}px; width: {WHITE_KEY_WIDTH - 1}px; height: {WHITE_KEY_HEIGHT}px;"
					title="{key.name}{getOctave(key.id)} — #{key.id} — {key.frequency}Hz"
				>
					{#if key.name === 'C'}
						<span
							class="pointer-events-none absolute bottom-2 left-0 right-0 text-center text-[9px] text-slate-400"
						>
							C{getOctave(key.id)}
						</span>
					{/if}
				</button>
			{/each}

			<!-- Black keys (rendered on top) -->
			{#each blackKeys as key}
				<button
					onclick={() => playNote(key)}
					class="absolute top-0 z-10 select-none rounded-b-md transition-colors {activeId === key.id
						? 'bg-indigo-600'
						: 'bg-slate-900 hover:bg-slate-700 dark:bg-zinc-900 dark:hover:bg-zinc-700'}"
					style="left: {key.x}px; width: {BLACK_KEY_WIDTH}px; height: {BLACK_KEY_HEIGHT}px;"
					title="{key.name}{getOctave(key.id)} — #{key.id} — {key.frequency}Hz"
				>
				</button>
			{/each}
		</div>
	</div>

	<p class="mt-2 text-xs {recording ? 'text-red-500' : 'text-slate-400'}">
		{recording
			? '● Recording — click keys to add notes to your song (max 16)'
			: 'Click keys to play. Enable Record to build your song. Click a queued note to remove it.'}
	</p>
</div>
