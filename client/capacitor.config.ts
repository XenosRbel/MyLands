import type { CapacitorConfig } from "@capacitor/cli";

// Android packaging of the web client (see docs/spec/PLATFORM_REQUIREMENTS.md).
// The WebView loads the bundled build from https://localhost, so the game
// server must allow that origin (RF_ALLOWED_ORIGINS) and the client must be
// built with VITE_API_BASE pointing at the public API host.
const config: CapacitorConfig = {
  appId: "com.realmforge.ashencrowns",
  appName: "Ashen Crowns",
  webDir: "dist",
  android: { allowMixedContent: false },
  server: { androidScheme: "https" },
};
export default config;
