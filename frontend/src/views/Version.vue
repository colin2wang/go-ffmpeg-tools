<template>
  <div class="page">
    <div class="card">
      <div class="card-title">{{ t('version.title') }}</div>

      <p v-if="error" class="tip-warn">{{ error }}</p>
      <p v-else-if="!info" class="muted-hint">{{ t('version.loading') }}</p>

      <template v-else>
        <table class="version-table">
          <tbody>
            <tr>
              <td class="label">{{ t('version.version') }}</td>
              <td>{{ info.version || '-' }}</td>
            </tr>
            <tr>
              <td class="label">{{ t('version.path') }}</td>
              <td class="path" :title="info.path">{{ info.path }}</td>
            </tr>
            <tr v-if="info.builtWith">
              <td class="label">{{ t('version.builtWith') }}</td>
              <td>{{ info.builtWith }}</td>
            </tr>
            <tr v-if="info.config">
              <td class="label">{{ t('version.config') }}</td>
              <td><pre class="version-config">{{ info.config }}</pre></td>
            </tr>
          </tbody>
        </table>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetFFmpegVersion } from '../../wailsjs/go/main/App'
import { ffmpeg } from '../../wailsjs/go/models'

const { t } = useI18n()
const info = ref<ffmpeg.VersionInfo | null>(null)
const error = ref('')

onMounted(async () => {
  try {
    info.value = await GetFFmpegVersion()
  } catch (e) {
    error.value = String(e)
  }
})
</script>

<style lang="scss" scoped>
.version-table {
  width: 100%;
  border-collapse: collapse;

  td {
    padding: 8px 12px 8px 0;
    vertical-align: top;
    word-break: break-all;
  }

  .label {
    width: 96px;
    white-space: nowrap;
    opacity: 0.65;
  }

  .version-config {
    margin: 0;
    white-space: pre-wrap;
    font-family: inherit;
    font-size: 12px;
    opacity: 0.85;
  }
}
</style>
