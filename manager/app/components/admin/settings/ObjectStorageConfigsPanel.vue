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
  isR2,
  isLocal,
  configHints,
  drawerTitle,
  drawerSubtitle,
  openCreate,
  openEdit,
  closeDrawer,
  goToConnect,
  backToVendor,
  submit,
  removeCurrent,
} = useObjectStorageConfigEditor();

const theme = (id: string) => vendorTheme(id);
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <AgentServiceHubPage
      v-model:page="page"
      v-model:page-size="pageSize"
      :lead="t('agentHub.storageLead')"
      :active-count="activeCount"
      :add-label="t('agentHub.addStorage')"
      :empty-title="t('agentHub.emptyTitle')"
      :empty-body="t('agentHub.emptyStorageBody')"
      :add-first-label="t('agentHub.addFirstStorage')"
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

        <template v-if="isR2">
          <AgentField :label="t('fields.accessKeyId')" required>
            <AdminInput v-model="form.apiKey" type="password" autocomplete="off" />
          </AgentField>
          <AgentField :label="t('fields.secretAccessKey')" required>
            <AdminInput v-model="form.secretKey" type="password" autocomplete="off" />
          </AgentField>
          <AgentField :label="t('fields.s3Endpoint')" required>
            <AdminInput v-model="form.baseUrl" />
          </AgentField>
          <AgentField :label="t('fields.bucketName')" required>
            <AdminInput v-model="form.bucket" />
          </AgentField>
          <AgentField :label="t('fields.publicAccessUrl')" required>
            <AdminInput v-model="form.publicBaseUrl" />
          </AgentField>
        </template>

        <template v-else-if="!isLocal">
          <AgentField :label="t('fields.endpointBaseUrl')">
            <AdminInput v-model="form.baseUrl" />
          </AgentField>
          <AgentField :label="t('fields.publicDomain')">
            <AdminInput v-model="form.publicBaseUrl" />
          </AgentField>
          <AgentField label="Access Key / API Key">
            <AdminInput v-model="form.apiKey" type="password" autocomplete="off" />
          </AgentField>
          <AgentField label="Secret Key">
            <AdminInput v-model="form.secretKey" type="password" autocomplete="off" />
          </AgentField>
          <AgentField :label="t('fields.bucket')">
            <AdminInput v-model="form.bucket" />
          </AgentField>
          <AgentField :label="t('common.region')">
            <AdminInput v-model="form.region" />
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
            <AgentField :label="t('fields.extJson')" :hint="configHints[form.provider]">
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
