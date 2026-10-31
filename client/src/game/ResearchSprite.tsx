/** Raster-only art. Every technology level has a dedicated painted WebP sprite. */
export function researchSpriteStage(level: number): number {
  return Math.max(1, Math.min(10, Math.trunc(Number.isFinite(level) ? level : 1)));
}

export function researchSpriteUrl(tech: string, level: number): string {
  if (!/^[a-z0-9_]+$/.test(tech)) return "/ui/research/heroics-1.webp";
  return `/ui/research/${tech}-${researchSpriteStage(level)}.webp`;
}

export function ResearchSprite({ tech, level, size = 46, className = "" }: {
  tech: string; level: number; size?: number; className?: string;
}) {
  return <img className={"research-sprite " + className}
    src={researchSpriteUrl(tech,level)}
    width={size} height={size} alt="" loading="lazy" decoding="async" draggable={false} />;
}
