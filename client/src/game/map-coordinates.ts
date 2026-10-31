/** Convert a pointer position to an integer tile in a CSS-scaled map. */
export function mapPointerToTile(
  clientX: number,
  clientY: number,
  bounds: { left: number; top: number; width: number; height: number },
  origin: { x: number; y: number },
  view: number,
): { x: number; y: number } {
  const x = Math.floor((clientX - bounds.left) / (bounds.width / view));
  const y = Math.floor((clientY - bounds.top) / (bounds.height / view));
  return { x: origin.x + x, y: origin.y + y };
}

/** Calculate drag offsets in tile coordinates using rendered dimensions. */
export function mapDragOffset(
  deltaX: number, deltaY: number,
  renderedWidth: number, renderedHeight: number,
  view: number,
): { x: number; y: number } {
  return {
    x: Math.round(deltaX / (renderedWidth / view)),
    y: Math.round(deltaY / (renderedHeight / view)),
  };
}
