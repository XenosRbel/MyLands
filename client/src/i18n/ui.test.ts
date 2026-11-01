import { describe, expect, it } from "vitest";
import { makeT, supportedLocales } from "./ui";

describe("game interface localization", () => {
  it("provides meaningful labels in every supported locale", () => {
    for (const locale of supportedLocales) {
      const t = makeT(locale);
      for (const key of ["city", "map", "army", "soundOn", "soundOff", "reducedMotion"] as const) {
        expect(t(key).length).toBeGreaterThan(0);
      }
    }
  });

  it("keeps audio labels distinct", () => {
    for (const locale of supportedLocales) {
      expect(makeT(locale)("soundOn")).not.toBe(makeT(locale)("soundOff"));
    }
  });
});
