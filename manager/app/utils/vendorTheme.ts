/** 服务商卡片视觉（缩写 + 渐变色） */
export type VendorTheme = {
  abbr: string;
  gradient: string;
  ring: string;
};

const THEMES: Record<string, VendorTheme> = {
  openai: { abbr: "AI", gradient: "from-emerald-500 to-teal-400", ring: "ring-emerald-500/30" },
  azure_openai: { abbr: "Az", gradient: "from-sky-500 to-blue-600", ring: "ring-sky-500/30" },
  deepseek: { abbr: "DS", gradient: "from-blue-600 to-indigo-500", ring: "ring-blue-500/30" },
  zhipu: { abbr: "智", gradient: "from-violet-500 to-purple-500", ring: "ring-violet-500/30" },
  moonshot: { abbr: "KM", gradient: "from-neutral-700 to-neutral-900", ring: "ring-neutral-500/30" },
  siliconflow: { abbr: "SF", gradient: "from-orange-500 to-amber-400", ring: "ring-orange-500/30" },
  openrouter: { abbr: "OR", gradient: "from-fuchsia-500 to-pink-500", ring: "ring-fuchsia-500/30" },
  groq: { abbr: "GQ", gradient: "from-lime-500 to-green-500", ring: "ring-lime-500/30" },
  together: { abbr: "TG", gradient: "from-cyan-500 to-blue-500", ring: "ring-cyan-500/30" },
  ollama: { abbr: "Ol", gradient: "from-stone-600 to-stone-800", ring: "ring-stone-500/30" },
  nvidia_nim: { abbr: "NV", gradient: "from-green-600 to-lime-500", ring: "ring-green-500/30" },
  mistral: { abbr: "Mi", gradient: "from-rose-500 to-orange-400", ring: "ring-rose-500/30" },
  claude: { abbr: "Cl", gradient: "from-amber-600 to-orange-500", ring: "ring-amber-500/30" },
  gemini: { abbr: "Ge", gradient: "from-blue-500 to-violet-500", ring: "ring-blue-500/30" },
  custom_openai: { abbr: "+", gradient: "from-primary-500 to-accent-500", ring: "ring-primary-500/30" },
};

export function vendorTheme(vendorId: string): VendorTheme {
  return (
    THEMES[vendorId] ?? {
      abbr: vendorId.slice(0, 2).toUpperCase(),
      gradient: "from-primary-500 to-accent-400",
      ring: "ring-primary-500/25",
    }
  );
}
