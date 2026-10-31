/** Original technology seals for Ashen Crowns. They are a visual layer only.
 * Technology rules, levels and prerequisites remain server authoritative.
 */
import { useId } from "react";

type Point = { cx: number; cy: number; r: number };
type Motif = { primary: string; secondary: string; accent: string; points: Point[] };

const motifs: Record<string, Motif> = {
  forestry: { primary:"M24 50L40 17L56 50H24ZM18 59L40 29L62 59H18Z", secondary:"M40 46V70M25 59L40 51L56 59",accent:"#71925d",points:[{cx:40,cy:13,r:3}]},
  agriculture: {primary:"M39 64V19M39 30Q18 13 21 32Q27 41 39 40M39 42Q59 18 60 33Q56 47 39 49M39 55Q20 41 18 55Q27 65 39 64",secondary:"M39 22L42 17M40 67L27 72",accent:"#c1ab66",points:[{cx:38,cy:13,r:3}]},
  masonry: {primary:"M18 28H61V42H18ZM25 45H67V59H25ZM14 62H61V72H14Z",secondary:"M40 28V42M45 45V59M35 62V72",accent:"#a5aca4",points:[]},
  metallurgy: {primary:"M19 54H61L54 68H29ZM29 31L40 13L52 31L48 49H33Z",secondary:"M22 55H58M41 36L35 47M43 35L49 46",accent:"#c78c65",points:[{cx:40,cy:24,r:4}]},
  exchequer: {primary:"M19 35Q19 21 40 20Q61 21 61 35V52Q61 65 40 65Q19 65 19 52ZM30 31H51V56H30Z",secondary:"M40 28V60M32 38Q48 30 50 42Q40 47 31 49Q33 57 50 52",accent:"#d0ac66",points:[]},
  husbandry: {primary:"M17 40L25 19L36 30H44L55 19L64 41L58 57L49 65H31L21 56Z",secondary:"M28 46H33M48 46H53M34 57Q40 63 47 57",accent:"#b6a07a",points:[{cx:30,cy:44,r:2},{cx:51,cy:44,r:2}]},
  tactics: {primary:"M22 21L40 13L60 22V40Q60 59 40 69Q20 59 20 40Z",secondary:"M29 29L51 51M51 29L29 51M40 20V61",accent:"#a1b29b",points:[]},
  armour: {primary:"M24 26L37 18H44L56 26L62 41L53 51L49 69H31L27 51L18 41Z",secondary:"M31 30H49M30 45H50M35 56H45",accent:"#aab6b3",points:[]},
  logistics: {primary:"M19 50L27 28H57L65 50L57 63H28ZM27 28L40 17L57 28",secondary:"M25 50H59M35 28V63M47 28V63",accent:"#a8a77b",points:[{cx:28,cy:67,r:4},{cx:57,cy:67,r:4}]},
  espionage: {primary:"M12 41Q40 11 68 41Q40 70 12 41ZM27 41A13 13 0 1 0 53 41A13 13 0 1 0 27 41",secondary:"M40 30L40 52M30 41H50",accent:"#9daab6",points:[{cx:40,cy:41,r:5}]},
  beastlore: {primary:"M22 52L15 36L24 24L34 29L40 18L46 29L59 24L67 36L57 52L53 65H28Z",secondary:"M29 43L34 47M47 47L53 43M36 57H46",accent:"#ba9976",points:[{cx:25,cy:33,r:3},{cx:55,cy:33,r:3}]},
  engineering: {primary:"M40 14L49 26L64 27L60 42L68 54L53 59L48 72L33 66L20 70L18 54L10 43L23 33L25 18Z",secondary:"M40 28A13 13 0 1 0 40 54A13 13 0 1 0 40 28",accent:"#b4a48e",points:[{cx:40,cy:41,r:5}]},
  arcana: {primary:"M40 12L61 40L40 69L19 40Z",secondary:"M40 12V69M19 40H61M29 27L52 53M52 27L29 53",accent:"#a79bc6",points:[{cx:40,cy:40,r:7}]},
  plunder: {primary:"M17 34L29 20L48 25L64 39L59 60L35 67L19 54Z",secondary:"M27 38L49 50M45 32L35 59M22 53L55 39",accent:"#bc886b",points:[]},
  heroics: {primary:"M40 12L49 30L68 33L54 47L57 67L40 57L23 67L26 47L12 33L31 30Z",secondary:"M40 24V53M30 40H50",accent:"#d9b37e",points:[{cx:40,cy:39,r:4}]},
  path_architect: {primary:"M17 59V29L27 22V35H35V15H45V35H53V22L64 29V59ZM31 59V45H49V59",secondary:"M17 29L40 22L64 29M18 63H64",accent:"#c5ad7a",points:[]},
  path_steward: {primary:"M18 54Q22 36 39 37L40 18Q62 21 63 43L55 64L30 65Z",secondary:"M40 21L40 69M24 50L50 36M40 55L61 43",accent:"#8fb589",points:[{cx:40,cy:17,r:3}]},
  path_warlord: {primary:"M24 18L55 57L47 65L17 29ZM56 19L25 57L33 65L64 29Z",secondary:"M40 20V68M28 49L51 50",accent:"#c79385",points:[]},
};

const fallback: Motif = motifs.heroics;

export function ResearchEmblem({ tech, level, size = 46, className = "" }: {
  tech: string; level: number; size?: number; className?: string;
}) {
  const id = useId().replace(/:/g, "");
  const emblem = motifs[tech] ?? fallback;
  const rank = Math.max(0, Math.min(10, Math.floor(level)));
  const stages = rank >= 8 ? 3 : rank >= 5 ? 2 : rank >= 2 ? 1 : 0;
  return (
    <svg className={"research-emblem " + className} width={size} height={size}
      viewBox="0 0 80 80" role="img" aria-label={tech.replaceAll("_"," ") + " " + rank}>
      <defs>
        <radialGradient id={id+"surface"} cx=".38" cy=".28" r=".85">
          <stop stopColor="#827055"/><stop offset=".55" stopColor="#4b3a2d"/><stop offset="1" stopColor="#211b1a"/>
        </radialGradient>
        <linearGradient id={id+"metal"} x1="0" y1="0" x2="1" y2="1">
          <stop stopColor="#f3dfb2"/><stop offset=".45" stopColor={emblem.accent}/><stop offset="1" stopColor="#55402d"/>
        </linearGradient>
      </defs>
      <path d="M40 4L68 14L74 40L66 65L40 77L14 65L6 40L12 14Z"
        fill={"url(#"+id+"surface)"} stroke="#1a1716" strokeWidth="3"/>
      <path d="M40 7L65 16L71 40L64 63L40 73L16 63L9 40L15 16Z"
        fill="none" stroke={"url(#"+id+"metal)"} strokeWidth={stages >= 2 ? 3.8 : 2}
        opacity={rank ? 1 : .58}/>
      <path d={emblem.primary} fill={emblem.accent} stroke="#342d26" strokeWidth="3" strokeLinejoin="round"/>
      <path d={emblem.secondary} fill="none" stroke="#f0d6a5" strokeWidth="3"
        strokeLinecap="round" strokeLinejoin="round" opacity=".9"/>
      {emblem.points.map((point,i) => <circle key={i} {...point} fill="#edddae"/>)}
      {stages >= 1 && <path d="M9 28L17 23M71 28L63 23M9 52L17 57M71 52L63 57"
        stroke="#d8bb84" strokeWidth="4" strokeLinecap="round"/>}
      {stages >= 2 && <path d="M22 9L26 4M54 9L58 4M22 71L26 76M54 71L58 76"
        stroke="#e4cc92" strokeWidth="4" strokeLinecap="round"/>}
      {stages >= 3 && <g fill="#f4dc97">
        <circle cx="40" cy="5" r="3"/><circle cx="75" cy="40" r="3"/>
        <circle cx="40" cy="75" r="3"/><circle cx="5" cy="40" r="3"/>
      </g>}
      <path d="M15 14Q25 8 35 8" fill="none" stroke="#fff0c5" strokeWidth="2" opacity=".3"/>
    </svg>
  );
}
