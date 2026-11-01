import { describe, expect, it } from "vitest";
import { webFontsUrl } from "./pack";

describe("webFontsUrl", () => {
  it("builds a Google Fonts URL for web builds", () => {
    const url = webFontsUrl(["Cinzel", "Source Sans 3"], "production");
    expect(url).toContain("https://fonts.googleapis.com/css2?");
    expect(url).toContain("family=Cinzel:wght@400;600;700");
    expect(url).toContain("family=Source%20Sans%203:wght@400;600;700");
  });

  it("never requests an external stylesheet in the Yandex build", () => {
    expect(webFontsUrl(["Cinzel"], "yandex")).toBeNull();
  });

  it("returns null when the pack declares no web fonts", () => {
    expect(webFontsUrl([], "production")).toBeNull();
  });
});
