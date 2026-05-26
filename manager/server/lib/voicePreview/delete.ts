import prisma from "../prisma";
import { deleteAssistantPreviewAudio } from "./storage";

export type DeletePreviewResult = {
  voiceRoleId: string;
  localDeleted: boolean;
  cloudKeys: string[];
};

export async function deleteVoiceRolePreview(voiceRoleId: string): Promise<DeletePreviewResult> {
  const role = await prisma.voiceRole.findUnique({ where: { id: voiceRoleId } });
  if (!role) {
    throw new Error("Voice role not found");
  }

  const hasPreview =
    Boolean(role.previewAudioUrl?.trim()) || Boolean(role.previewLocalFilename?.trim());
  if (!hasPreview) {
    throw new Error("Voice role has no preview audio");
  }

  const { localDeleted, cloudKeys } = await deleteAssistantPreviewAudio({
    previewAudioUrl: role.previewAudioUrl,
    previewLocalFilename: role.previewLocalFilename,
  });

  await prisma.voiceRole.update({
    where: { id: voiceRoleId },
    data: {
      previewAudioUrl: null,
      previewLocalFilename: null,
      previewGeneratedAt: null,
    },
  });

  return { voiceRoleId, localDeleted, cloudKeys };
}
