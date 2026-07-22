<script setup lang="ts">
import { onMounted } from 'vue'
import { useFriendStore } from '@/stores/friend'
import { useAuthStore } from '@/stores/auth'
import { getFriendApplyListApi, handleFriendRequestApi } from '@/api/friend'
import { getAvatarUrl } from '@/utils/avatar'

const friendStore = useFriendStore()
const auth = useAuthStore()

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
  <div class="notify-panel">
    <div class="panel-header">
      <h3>通知中心</h3>
    </div>

    <div class="notify-content">
      <div class="section-title">好友申请</div>
      <div v-if="friendStore.applyList.length === 0" class="empty-tip">暂无好友申请</div>
      <div v-for="apply in friendStore.applyList" :key="apply.id" class="notify-card">
        <div class="card-avatar">
          <img v-if="getAvatarUrl(apply.avatar)" :src="getAvatarUrl(apply.avatar)" class="avatar-img" />
          <span v-else>{{ apply.friend_name?.[0] || '?' }}</span>
        </div>
        <div class="card-info">
          <div class="card-name">{{ apply.friend_name }}</div>
          <div class="card-signature">{{ apply.signature || '暂无签名' }}</div>
          <div class="card-actions">
            <button class="btn-accept" @click="handleRequest(apply.id, apply.id, true)">同意</button>
            <button class="btn-reject" @click="handleRequest(apply.id, apply.id, false)">拒绝</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.notify-panel { display: flex; flex-direction: column; height: 100%; background: #fff; }
.notify-content { flex: 1; overflow-y: auto; padding: 8px 16px; }
.section-title { font-size: 13px; font-weight: 600; color: #666; padding: 12px 0 8px; }
.notify-card {
  display: flex; gap: 14px; padding: 14px 16px;
  background: #f8f9ff; border-radius: 10px; margin-bottom: 10px;
  border: 1px solid #eef0f7;
  transition: box-shadow 0.15s;
}
.notify-card:hover { box-shadow: 0 2px 8px rgba(102,126,234,0.12); }
.card-avatar {
  width: 48px; height: 48px; border-radius: 12px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: #fff; display: flex; align-items: center; justify-content: center;
  font-size: 18px; font-weight: 600; flex-shrink: 0; overflow: hidden;
}
.card-avatar .avatar-img { width: 100%; height: 100%; object-fit: cover; }
.card-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px; }
.card-name { font-size: 15px; font-weight: 600; color: #333; }
.card-signature { font-size: 12px; color: #999; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.card-actions { display: flex; gap: 8px; margin-top: 6px; }
.btn-accept, .btn-reject {
  padding: 6px 20px; border: none; border-radius: 6px;
  font-size: 13px; font-weight: 500; cursor: pointer; transition: background 0.15s;
}
.btn-accept { background: #667eea; color: #fff; }
.btn-accept:hover { background: #5a6fd6; }
.btn-reject { background: #f0f0f0; color: #666; }
.btn-reject:hover { background: #e0e0e0; color: #e74c3c; }
</style>
