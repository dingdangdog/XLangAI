// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  app: {
    head: {
      title: "小浪AI 运营后台",
      link: [
        {
          rel: "stylesheet",
          href: "https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap",
        },
      ],
    },
  },
  ssr: false,
  devtools: { enabled: true },
  modules: ["@nuxtjs/tailwindcss", "@pinia/nuxt", "@nuxtjs/i18n"],
  css: ["~/assets/css/themes.css", "~/assets/css/base.css"],
  i18n: {
    strategy: "prefix_except_default",
    defaultLocale: "zh",
    langDir: "locales",
    locales: [
      { code: "zh", name: "简体中文", language: "zh-CN", file: "zh.json" },
      { code: "en", name: "English", language: "en-US", file: "en.json" },
      { code: "ja", name: "日本語", language: "ja-JP", file: "ja.json" },
    ],
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: "xlangai_manager_i18n_redirected",
      redirectOn: "root",
      fallbackLocale: "zh",
    },
  },
  vite: {
    optimizeDeps: {
      include: ["dayjs/plugin/*.js"],
    },
  },
  runtimeConfig: {
    manager: {
      /** 启动时是否执行 Prisma 迁移；NUXT_MANAGER_DATABASE_AUTO_MIGRATE */
      databaseAutoMigrate: "true",
      /** 运营后台 JWT；NUXT_MANAGER_AUTH_SECRET */
      authSecret: "",
      /** 业务种子；NUXT_MANAGER_AUTO_SEED */
      autoSeed: "true",
      /** 联调测试账号；NUXT_MANAGER_TEST_ACCOUNT_SEED */
      testAccountSeed: "false",
      /** 首次管理员；NUXT_MANAGER_ADMIN_* */
      adminUsername: "",
      adminPassword: "",
      adminNickname: "",
      adminSeed: "true",
    },
    public: {
      /** 官网 / 服务器商店；NUXT_PUBLIC_OFFICIAL_HOME_URL */
      officialHomeUrl: "https://xlangai.com",
    },
  },
});
