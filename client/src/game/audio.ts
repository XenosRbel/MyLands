/** Original synthesized UI cues. No external audio assets are required. */
export type GameCue = "navigate" | "confirm" | "reward" | "warning";
let context: AudioContext | undefined;
let muted = true;
export function setGameAudioEnabled(enabled: boolean) { muted = !enabled; }
export function gameAudioEnabled() { return !muted; }
export function playGameCue(cue: GameCue) {
  if (muted || typeof window === "undefined") return;
  try {
    context ??= new AudioContext();
    if (context.state === "suspended") void context.resume();
    const notes: Record<GameCue, number[]> = {
      navigate: [420], confirm: [520, 780], reward: [523, 659, 880], warning: [220, 185],
    };
    notes[cue].forEach((frequency, index) => {
      const oscillator = context!.createOscillator();
      const gain = context!.createGain();
      const start = context!.currentTime + index * 0.065;
      oscillator.type = cue === "warning" ? "triangle" : "sine";
      oscillator.frequency.value = frequency;
      gain.gain.setValueAtTime(0.0001, start);
      gain.gain.exponentialRampToValueAtTime(0.035, start + 0.012);
      gain.gain.exponentialRampToValueAtTime(0.0001, start + 0.16);
      oscillator.connect(gain).connect(context!.destination);
      oscillator.start(start);
      oscillator.stop(start + 0.17);
    });
  } catch { /* Audio is optional. Browsers may restrict playback. */ }
}
