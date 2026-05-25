const DEFAULT_PREVIEW_TEMPLATES: Record<string, string> = {
  zh: "你好，我是{name}",
  yue: "你好，我係{name}",
  en: "Hello, I'm {name}",
  ja: "こんにちは、{name}です",
  ko: "안녕하세요, 저는 {name}입니다",
  es: "Hola, soy {name}",
  fr: "Bonjour, je suis {name}",
  de: "Hallo, ich bin {name}",
  pt: "Olá, eu sou {name}",
  it: "Ciao, sono {name}",
  ru: "Привет, я {name}",
  ar: "مرحباً، أنا {name}",
  hi: "नमस्ते, मैं {name} हूँ",
};

export function resolvePreviewSampleText(
  langCode: string,
  langTemplate: string | null | undefined,
  roleName: string,
): string {
  let tpl = (langTemplate ?? "").trim();
  if (!tpl) {
    tpl = DEFAULT_PREVIEW_TEMPLATES[langCode.trim().toLowerCase()] ?? "";
  }
  if (!tpl) {
    tpl = "Hello, I'm {name}";
  }
  const name = roleName.trim() || "AI";
  return tpl
    .replaceAll("{{voice_role_name}}", name)
    .replaceAll("{{name}}", name)
    .replaceAll("{name}", name);
}
