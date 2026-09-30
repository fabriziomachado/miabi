<script setup lang="ts">
import { computed, ref } from 'vue'
import { useComponents } from '@/composables/useComponents'

// The dismissible explanation under a Beta or Preview page; BetaBadge beside the title stays.
// Driven by internal/components, so promoting a component to Stable removes both.
const props = defineProps<{ component: string }>()

const { byId } = useComponents()
const info = computed(() => byId(props.component))
const early = computed(() => info.value?.stability === 'beta' || info.value?.stability === 'preview')

// Dismissal is per component and version: a new version while still in beta shows it again.
const storageKey = computed(() => `miabi.beta-dismissed.${props.component}`)
const dismissedVersion = ref(read(storageKey.value))

function read(key: string): string | null {
  try {
    return localStorage.getItem(key)
  } catch {
    return null
  }
}

function dismiss() {
  const v = info.value?.version ?? ''
  dismissedVersion.value = v
  try {
    localStorage.setItem(storageKey.value, v)
  } catch {
    // private mode or blocked storage: hidden for this page view only
  }
}

const visible = computed(() => early.value && dismissedVersion.value !== info.value?.version)
</script>

<template>
  <div v-if="visible && info" class="beta-banner" role="note">
    <span class="mdi mdi-flask-outline beta-icon"></span>
    <div class="beta-text">
      <strong>{{ $t(`beta.title.${info.stability}`, { name: info.name }) }}</strong>
      {{ $t(`beta.body.${info.stability}`) }}
      <a href="https://github.com/miabi-io/miabi/issues" target="_blank" rel="noopener noreferrer">{{ $t('beta.feedback') }}</a>
    </div>
    <button class="btn-icon btn-icon-muted" :title="$t('beta.dismiss')" :aria-label="$t('beta.dismiss')" @click="dismiss">
      <span class="mdi mdi-close"></span>
    </button>
  </div>
</template>

<style scoped>
.beta-banner {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 16px;
  padding: 10px 12px;
  border: 1px solid var(--warning-600);
  border-radius: var(--radius);
  background: var(--warning-50);
  color: var(--text-primary);
  font-size: 13px;
  line-height: 1.5;
}
.beta-icon {
  flex-shrink: 0;
  font-size: 18px;
  line-height: 1.2;
  color: var(--warning-600);
}
.beta-text {
  flex: 1;
  min-width: 0;
}
.beta-text a {
  white-space: nowrap;
}
</style>
