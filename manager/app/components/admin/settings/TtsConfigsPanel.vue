<script setup lang="ts">
import { vendorTheme } from "~/utils/vendorTheme";

const {
  t,
  loading,
  total,
  page,
  pageSize,
  activeCount,
  drawerOpen,
  setupPhase,
  setupStepNumber,
  editorMode,
  form,
  schemaFields,
  configSchema,
  dirty,
  saving,
  deleting,
  probing,
  lastResult,
  probeVoiceCode,
  regionRequired,
  tiles,
  vendorGridItems,
  activePreset,
  drawerTitle,
  drawerSubtitle,
  openCreate,
  openEdit,
  closeDrawer,
  goToConnect,
  backToVendor,
  submit,
  removeCurrent,
  runProbe,
} = useTtsConfigEditor();

const theme = (id: string) => vendorTheme(id);
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <AgentServiceHubPage
      v-model:page="page"
      v-model:page-size="pageSize"
      :lead="t('agentHub.ttsLead')"
      :active-count="activeCount"
      :add-label="t('agentHub.addTts')"
      :empty-title="t('agentHub.emptyTitle')"
      :empty-body="t('agentHub.emptyTtsBody')"
      :add-first-label="t('agentHub.addFirstTts')"
      :tiles="tiles"
      :loading="loading"
      :total="total"
      @create="openCreate"
      @edit="openEdit"
    />

    <AgentDrawer :open="drawerOpen" :title="drawerTitle" :subtitle="drawerSubtitle" wide @close="closeDrawer">
      <template #header-extra>
        <div v-if="editorMode === 'create'" class="mt-4">
          <AgentStepIndicator :current="setupStepNumber" :total="2" />
        </div>
      </template>

      <div v-if="setupPhase === 'vendor'">
        <AgentVendorGrid v-model="form.vendorPresetId" :items="vendorGridItems" />
      </div>

      <div v-else class="space-y-6">
        <div v-if="activePreset" class="flex items-center gap-3 rounded-2xl bg-surface-muted/60 p-4">
          <div
            class="flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br text-sm font-bold text-white"
            :class="theme(form.vendorPresetId).gradient"
          >
            {{ theme(form.vendorPresetId).abbr }}
          </div>
          <div class="min-w-0 flex-1">
            <p class="font-medium">{{ t(activePreset.labelKey) }}</p>
            <p class="text-xs text-muted">{{ t("agentHub.apiKeyHint") }}</p>
          </div>
          <button
            v-if="editorMode === 'create'"
            type="button"
            class="shrink-0 text-xs text-primary-600 hover:underline"
            @click="backToVendor"
          >
            {{ t("agentHub.changeVendor") }}
          </button>
        </div>

        <AgentField :label="t('serverCatalog.llm.displayName')">
          <AdminInput v-model="form.name" />
        </AgentField>

        <AgentField :label="t('fields.baseUrl')" :hint="t('pages.ttsConfigs.baseUrlHint')">
          <AdminInput v-model="form.baseUrl" />
        </AgentField>

        <AgentField :label="t('fields.apiKeySecret')" :hint="t('pages.ttsConfigs.defaultApiKeyHint')">
          <AdminInput v-model="form.apiKey" type="password" autocomplete="off" />
        </AgentField>

        <AgentField
          :label="t('common.region')"
          :required="regionRequired"
          :hint="regionRequired ? t('pages.ttsConfigs.regionRequired') : t('pages.ttsConfigs.regionOptional')"
        >
          <AdminInput v-model="form.region" placeholder="eastus / ap-guangzhou / us-east-1" />
        </AgentField>

        <AgentField :label="t('fields.modelCode')" :hint="t('pages.ttsConfigs.modelHint')">
          <AdminInput v-model="form.modelCode" />
        </AgentField>

        <label
          class="flex cursor-pointer items-center justify-between rounded-2xl border border-border/70 px-4 py-3.5 transition hover:border-primary-300/50"
        >
          <div>
            <p class="text-sm font-medium">{{ t("agentHub.enableConnection") }}</p>
            <p class="text-xs text-muted">{{ t("agentHub.enableConnectionHint") }}</p>
          </div>
          <input
            type="checkbox"
            class="h-5 w-5 rounded border-border text-primary-600"
            :checked="form.status === 'active'"
            @change="form.status = ($event.target as HTMLInputElement).checked ? 'active' : 'inactive'"
          />
        </label>

        <div v-if="editorMode === 'edit'" class="rounded-2xl border border-border/60 p-4">
          <AdminConfigProbePanel
            :loading="probing"
            :result="lastResult"
            :disabled="!form.id"
            :hint="t('serviceConfig.ttsProbeHint')"
            @probe="runProbe"
          />
          <AgentField class="mt-4" :label="t('serviceConfig.probeVoiceCode')" :hint="t('serviceConfig.probeVoiceHint')">
            <AdminInput v-model="probeVoiceCode" placeholder="alloy / en-US-JennyNeural" />
          </AgentField>
        </div>

        <details class="rounded-2xl border border-border/60">
          <summary class="cursor-pointer px-4 py-3 text-sm text-muted hover:text-foreground">
            {{ t("agentHub.advancedSettings") }}
          </summary>
          <div class="space-y-4 border-t border-border/50 px-4 py-4">
            <AdminConfigSchemaFields v-model="schemaFields" :schema="configSchema" />
            <AgentField :label="t('fields.extJson')">
              <AdminInput v-model="form.config" type="textarea" :rows="6" class="font-mono text-xs" />
            </AgentField>
            <AgentField :label="t('common.remark')">
              <AdminInput v-model="form.remark" type="textarea" :rows="2" />
            </AgentField>
          </div>
        </details>
      </div>

      <template #footer>
        <AgentHubDrawerFooter
          :editor-mode="editorMode"
          :setup-phase="setupPhase"
          :setup-step-number="setupStepNumber"
          :saving="saving"
          :deleting="deleting"
          :dirty="dirty"
          show-delete
          @close="closeDrawer"
          @back="backToVendor"
          @next="goToConnect"
          @save="submit"
          @delete="removeCurrent"
        />
      </template>
    </AgentDrawer>
  </div>
</template>
