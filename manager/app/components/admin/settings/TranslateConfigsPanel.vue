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
  dirty,
  saving,
  deleting,
  tiles,
  vendorGridItems,
  activePreset,
  isOpenAi,
  useLlmLink,
  llmConfigOptions,
  drawerTitle,
  drawerSubtitle,
  apiKeyLabel,
  apiSecretLabel,
  needsApiSecret,
  configHints,
  openCreate,
  openEdit,
  closeDrawer,
  goToConnect,
  backToVendor,
  submit,
  removeCurrent,
} = useTranslateConfigEditor();

const theme = (id: string) => vendorTheme(id);
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <AgentServiceHubPage
      v-model:page="page"
      v-model:page-size="pageSize"
      :lead="t('agentHub.translateLead')"
      :active-count="activeCount"
      :add-label="t('agentHub.addTranslate')"
      :empty-title="t('agentHub.emptyTitle')"
      :empty-body="t('agentHub.emptyTranslateBody')"
      :add-first-label="t('agentHub.addFirstTranslate')"
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
            <p v-if="activePreset.descriptionKey" class="text-xs text-muted">
              {{ t(activePreset.descriptionKey) }}
            </p>
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

        <template v-if="isOpenAi">
          <label
            class="flex cursor-pointer items-center gap-3 rounded-2xl border border-border/70 px-4 py-3.5 transition hover:border-primary-300/50"
          >
            <input v-model="useLlmLink" type="checkbox" class="h-5 w-5 rounded border-border text-primary-600" />
            <div>
              <p class="text-sm font-medium">{{ t("pages.translateConfigs.linkLlmRecommended") }}</p>
              <p class="text-xs text-muted">{{ t("pages.translateConfigs.llmSource") }}</p>
            </div>
          </label>

          <AgentField v-if="useLlmLink" :label="t('fields.linkedLlmShort')" required>
            <AdminSelect
              v-model="form.llmConfigId"
              :options="llmConfigOptions"
              :placeholder="t('pages.translateConfigs.selectActiveLlm')"
            />
            <p v-if="llmConfigOptions.length === 0" class="mt-1 text-xs text-muted">
              {{ t("pages.translateConfigs.createLlmFirst") }}
            </p>
          </AgentField>

          <template v-else>
            <AgentField :label="t('fields.baseUrl')" :hint="t('pages.translateConfigs.baseUrlHint')">
              <AdminInput v-model="form.baseUrl" :placeholder="t('pages.translateConfigs.baseUrlPlaceholder')" />
            </AgentField>
            <AgentField :label="t('fields.apiKey')" :hint="t('pages.translateConfigs.apiKeyRequired')">
              <AdminInput v-model="form.apiKey" type="password" autocomplete="off" />
            </AgentField>
            <AgentField :label="t('fields.modelCode')" required>
              <AdminInput v-model="form.modelCode" placeholder="gpt-4o-mini" />
            </AgentField>
          </template>

          <AgentField
            v-if="useLlmLink"
            :label="`${t('fields.modelCode')} (${t('common.optional')})`"
            :hint="t('pages.translateConfigs.modelOverrideHint')"
          >
            <AdminInput v-model="form.modelCode" :placeholder="t('pages.translateConfigs.modelOverride')" />
          </AgentField>
        </template>

        <template v-else>
          <AgentField
            v-if="form.protocol === 'azure_translator' || form.protocol === 'deepl'"
            :label="t('fields.baseUrl')"
            :hint="t('pages.translateConfigs.customEndpointHint')"
          >
            <AdminInput v-model="form.baseUrl" :placeholder="t('pages.llmConfigs.optionalPlaceholder')" />
          </AgentField>
          <AgentField :label="apiKeyLabel(form.protocol)">
            <AdminInput v-model="form.apiKey" type="password" autocomplete="off" />
          </AgentField>
          <AgentField v-if="needsApiSecret(form.protocol)" :label="apiSecretLabel(form.protocol)">
            <AdminInput v-model="form.apiSecret" type="password" autocomplete="off" />
          </AgentField>
        </template>

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

        <details class="rounded-2xl border border-border/60">
          <summary class="cursor-pointer px-4 py-3 text-sm text-muted hover:text-foreground">
            {{ t("agentHub.advancedSettings") }}
          </summary>
          <div class="space-y-4 border-t border-border/50 px-4 py-4">
            <AgentField :label="t('fields.extJson')" :hint="configHints[form.protocol]">
              <AdminInput v-model="form.config" type="textarea" :rows="4" class="font-mono text-xs" />
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
