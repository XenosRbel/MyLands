import { describe, expect, it } from "vitest";
import { renderToStaticMarkup } from "react-dom/server";
import { ResearchSprite, researchSpriteStage, researchSpriteUrl } from "./ResearchSprite";
import { Icon } from "../ui/kit";

describe("raster-only artwork", () => {
  it("updates research sprite artwork for every progression stage", () => {
    expect(researchSpriteStage(0)).toBe(1);
    expect(researchSpriteStage(3)).toBe(3);
    expect(researchSpriteStage(4)).toBe(4);
    expect(researchSpriteStage(7)).toBe(7);
    expect(researchSpriteStage(10)).toBe(10);
    expect(researchSpriteStage(999)).toBe(10);
    expect(researchSpriteUrl("path_architect",8)).toBe("/ui/research/path_architect-8.webp");
  });
  it("never embeds an SVG in research or UI controls", () => {
    const art = renderToStaticMarkup(<ResearchSprite tech="forestry" level={9} />);
    const icon = renderToStaticMarkup(<Icon name="clock" title="Time" />);
    expect(art).toContain('forestry-9.webp');
    expect(icon).toContain('/ui/glyphs/clock.webp');
    expect(art+icon).not.toContain("<svg");
  });
  it("rejects untrusted icon paths", () => {
    expect(researchSpriteUrl("../x.svg",1)).toBe("/ui/research/heroics-1.webp");
  });
});
