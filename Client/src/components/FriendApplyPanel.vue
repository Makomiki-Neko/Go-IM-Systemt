<script setup lang="ts">
import { onMounted } from 'vue'
import { useFriendStore } from '@/stores/friend'
import { useAuthStore } from '@/stores/auth'
import { getFriendApplyListApi, handleFriendRequestApi } from '@/api/friend'

const friendStore = useFriendStore()
const auth = useAuthStore()

const emit = defineEmits<{
  close: []
}>()

onMounted(loadApplies)

async function loadApplies() {
  try {
    const { data } = await getFriendApplyListApi(1, 200)
    if (data.list) friendStore.setApplyList(data.list)
  } catch {}
}

async function handleRequest(requestId: string, senderId: string, accept: boolean) {
  try {
    await handleFriendRequestApi({
      request_id: requestId,
      accept,
      sender_id: senderId,
      user_id: auth.uid!,
    })
    friendStore.removeApply(requestId)
  } catch {}
}
</script>

<template>
  <div class="apply-panel">
    <div class="panel-header">
      <h3>好友申请</h3>
      <button class="btn-close-inline" @click="emit('close')">&#x2715;</button>
    </div>
    <div class="apply-list">
      <div v-if="friendStore.applyList.length === 0" class="empty-tip">暂无好友申请</div>
      <div v-for="apply in friendStore.applyList" :key="apply.id" class="apply-item">
        <div class="apply-avatar">{{ apply.friend_name?.[0] || '?' }}</div>
        <div class="apply-info">
          <div class="apply-name">{{ apply.friend_name }}</div>
          <div class="apply-msg">{{ apply.signature || '' }}</div>
        </div>
        <div class="apply-actions">
          <button class="btn-small btn-accept" @click="handleRequest(apply.id, apply.id, true)">同意</button>
          <button class="btn-small btn-danger" @click="handleRequest(apply.id, apply.id, false)">拒绝</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.btn-close-inline { background: none; border: none; font-size: 16px; cursor: pointer; color: #999; padding: 2px 8px; }
.btn-close-inline:hover { color: #333; }
</style>
