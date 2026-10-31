import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import { detectPlatform } from "./platform/adapter";
import "./styles.css";
import "./game/experience.css";
import "./game/realm-theme.css";

try {
  document.documentElement.dataset.motion = localStorage.getItem("rf.motion") ?? "";
} catch {
  /* storage unavailable */
}

const platform = detectPlatform();
platform.init().finally(() => {
  createRoot(document.getElementById("root")!).render(
    <StrictMode>
      <App platform={platform} />
    </StrictMode>,
  );
});
