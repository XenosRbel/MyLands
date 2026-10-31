import { describe, expect, it } from "vitest";
import { renderToStaticMarkup } from "react-dom/server";
import { ResearchEmblem } from "./ResearchEmblem";

describe("research illustrated technology seals", () => {
  it("uses unique silhouettes for different technologies", () => {
    const timber = renderToStaticMarkup(<ResearchEmblem tech="forestry" level={3}/>);
    const agriculture = renderToStaticMarkup(<ResearchEmblem tech="agriculture" level={3}/>);
    expect(timber).toContain("M24 50L40 17");
    expect(agriculture).toContain("M39 64V19");
    expect(timber).not.toBe(agriculture);
  });

  it("changes the visible ornament when the same technology advances", () => {
    const novice = renderToStaticMarkup(<ResearchEmblem tech="arcana" level={1}/>);
    const veteran = renderToStaticMarkup(<ResearchEmblem tech="arcana" level={9}/>);
    expect(novice).not.toContain('cx="75" cy="40"');
    expect(veteran).toContain('cx="75" cy="40"');
  });

  it("does not render the artwork of an unrelated building", () => {
    const portrait = renderToStaticMarkup(<ResearchEmblem tech="engineering" level={5}/>);
    expect(portrait).toContain("<svg");
    expect(portrait).not.toContain("<img");
    expect(portrait).not.toContain("/buildings/");
  });
});
