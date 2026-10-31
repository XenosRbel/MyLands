import { describe, expect, it } from "vitest";
import { mapDragOffset, mapPointerToTile } from "./map-coordinates";

describe("scaled world map interactions", () => {
  it("selects the same location when the map is CSS-scaled", () => {
    const origin = { x: 40, y: 60 };
    const rect = { left: 150, top: 100, width: 750, height: 750 };
    expect(mapPointerToTile(150 + 375, 100 + 375, rect, origin, 15)).toEqual({ x: 47, y: 67 });
    const smaller = { left: 150, top: 100, width: 450, height: 450 };
    expect(mapPointerToTile(150 + 225, 100 + 225, smaller, origin, 15)).toEqual({ x: 47, y: 67 });
  });

  it("uses independent axes for non-square display bounds", () => {
    const rect = { left: 10, top: 20, width: 600, height: 450 };
    expect(mapPointerToTile(10 + 80, 20 + 90, rect, { x: 0, y: 0 }, 15)).toEqual({ x: 2, y: 3 });
  });

  it("translates drags into equivalent tile offsets after scaling", () => {
    expect(mapDragOffset(100, -50, 750, 750, 15)).toEqual({ x: 2, y: -1 });
    expect(mapDragOffset(60, -30, 450, 450, 15)).toEqual({ x: 2, y: -1 });
  });
});
