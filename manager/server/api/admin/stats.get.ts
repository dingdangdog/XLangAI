import prisma from "../../lib/prisma";

export default defineEventHandler(async () => {
  const [
    users,
    languages,
    llmConfigs,
    sttConfigs,
    ttsConfigs,
    voiceRoles,
    promptTemplates,
    conversations,
    messages,
    tiers,
  ] = await Promise.all([
    prisma.user.count({ where: { deletedAt: null } }),
    prisma.language.count(),
    prisma.sysLlmServiceConfig.count(),
    prisma.sysSttServiceConfig.count(),
    prisma.ttsServiceConfig.count(),
    prisma.voiceRole.count(),
    prisma.promptTemplate.count(),
    prisma.conversation.count({ where: { deletedAt: null } }),
    prisma.message.count({ where: { deletedAt: null } }),
    prisma.membershipTier.count(),
  ]);

  return {
    users,
    languages,
    llmConfigs,
    sttConfigs,
    ttsConfigs,
    voiceRoles,
    promptTemplates,
    conversations,
    messages,
    tiers,
  };
});
