import { describe, expect, it } from "vitest";
import { gameAudioEnabled, playGameCue, setGameAudioEnabled } from "./audio";
describe("game audio settings", () => {
  it("starts muted and supports opt-in", () => {
    setGameAudioEnabled(false);
    expect(gameAudioEnabled()).toBe(false);
    playGameCue("navigate");
    setGameAudioEnabled(true);
    expect(gameAudioEnabled()).toBe(true);
    setGameAudioEnabled(false);
  });
});
