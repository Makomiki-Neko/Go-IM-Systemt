<script setup lang="ts">
import { ref } from 'vue'
import type { FriendInfo } from '@/types'
import { useFriendStore } from '@/stores/friend'
import { deleteFriendApi } from '@/api/friend'
import { getAvatarUrl } from '@/utils/avatar'

const props = defineProps<{
  friend: FriendInfo
}>()

const emit = defineEmits<{
  startChat: [userId: string]
  close: []
}>()

const friendStore = useFriendStore()
const showRemarkInput = ref(false)
const newRemark = ref(props.friend.remark || props.friend.friend_name)

async function delFriend() {
  try {
    await deleteFriendApi({ friend_id: props.friend.id })
    friendStore.removeFriend(props.friend.id)
    emit('close')
  } catch {}
}
</script>

<template>
  <div class="friend-detail">
    <div class="detail-header">
      <h3>好友信息</h3>
      <button class="btn-close" @click="emit('close')">✕</button>
    </div>
    <div class="detail-content">
      <div class="detail-avatar-large">
        <img v-if="getAvatarUrl(friend.avatar)" :src="getAvatarUrl(friend.avatar)" class="avatar-img" />
        <span v-else>{{ friend.remark?.[0] || friend.friend_name?.[0] || '?' }}</span>
      </div>
      <h2 class="detail-name">{{ friend.remark || friend.friend_name }}</h2>
      <div class="detail-field">
        <label>账号</label>
        <span>{{ friend.friend_name }}</span>
      </div>
      <div class="detail-field">
        <label>备注</label>
        <template v-if="showRemarkInput">
          <div class="remark-edit">
            <input v-model="newRemark" />
            <button class="btn-small" @click="showRemarkInput = false">保存</button>
          </div>
        </template>
        <span v-else class="remark-text" @click="showRemarkInput = true">{{ friend.remark || '无' }}</span>
      </div>
      <div class="detail-field">
        <label>签名</label>
        <span>{{ friend.signature || '未设置' }}</span>
      </div>
      <div class="detail-field">
        <label>性别</label>
        <span>{{ friend.gender === 'male' ? '男' : friend.gender === 'female' ? '女' : '保密' }}</span>
      </div>
      <div class="detail-field">
        <label>在线状态</label>
        <span :class="friend.online_statu ? 'online-text' : ''">
          {{ friend.online_statu ? '在线' : '离线' }}
        </span>
      </div>

      <div class="detail-actions">
        <button class="btn-primary" @click="emit('startChat', friend.id)">发消息</button>
        <button class="btn-secondary btn-danger-outline" @click="delFriend">删除好友</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.friend-detail { display: flex; flex-direction: column; height: 100%; background: #fff; }
.detail-header { display: flex; align-items: center; justify-content: space-between; padding: 14px 20px; border-bottom: 1px solid #e8e8e8; }
.detail-header h3 { font-size: 16px; }
.btn-close { background: none; border: none; font-size: 18px; cursor: pointer; color: #999; padding: 4px 8px; border-radius: 4px; }
.btn-close:hover { background: #f0f0f0; color: #333; }
.detail-content { flex: 1; overflow-y: auto; padding: 24px 20px; }
.detail-avatar-large { width: 72px; height: 72px; border-radius: 12px; background: #667eea; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 28px; font-weight: 600; margin: 0 auto 12px; }
.detail-name { text-align: center; font-size: 18px; margin-bottom: 24px; }
.detail-field { margin-bottom: 14px; }
.detail-field label { display: block; font-size: 12px; color: #999; margin-bottom: 2px; }
.detail-field span { font-size: 14px; color: #333; }
.remark-text { cursor: pointer; border-bottom: 1px dashed #ddd; }
.remark-edit { display: flex; gap: 6px; }
.remark-edit input { flex: 1; padding: 6px 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 13px; outline: none; }
.online-text { color: #2ecc71; }
.detail-actions { display: flex; gap: 10px; margin-top: 24px; }
.detail-actions button { flex: 1; text-align: center; }
.btn-danger-outline { color: #e74c3c; border-color: #e74c3c; }
.btn-danger-outline:hover { background: #fde8e8; }
</style>
