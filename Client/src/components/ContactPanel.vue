<script setup lang="ts">
import { ref, computed } from 'vue'
import { useFriendStore } from '@/stores/friend'
import { searchUserApi, sendFriendRequestApi } from '@/api/friend'
import { getAvatarUrl } from '@/utils/avatar'
import type { FriendInfo } from '@/types'

const emit = defineEmits<{
  showDetail: [friend: FriendInfo]
  startChat: [userId: string]
}>()

const friendStore = useFriendStore()
const searchName = ref('')
const searchResults = ref<FriendInfo[]>([])
const sendingIds = ref<Set<string>>(new Set())
const sentIds = ref<Set<string>>(new Set())

const friendIds = computed(() => new Set(friendStore.friendList.map((f) => f.id)))

async function handleSearch() {
  if (!searchName.value.trim()) return
  try {
    const { data } = await searchUserApi(searchName.value.trim())
    searchResults.value = data.list || []
  } catch {}
}

async function addFriend(friendId: string) {
  sendingIds.value.add(friendId)
  try {
    await sendFriendRequestApi({
      friend_id: friendId,
      msg: '\u4F60\u597D\uFF0C\u6211\u662F' + (localStorage.getItem('account') || ''),
    })
    sentIds.value.add(friendId)
  } catch {} finally {
    sendingIds.value.delete(friendId)
  }
}

function openNotify() {
  window.dispatchEvent(new CustomEvent('im:showNotify'))
}
</script>

<template>
  <div class="contact-panel">
    <div class="panel-header">
      <h3>好友列表</h3>
    </div>

    <div class="search-bar">
      <input v-model="searchName" placeholder="搜索用户..." @keydown.enter="handleSearch" />
      <button class="btn-small" @click="handleSearch">搜索</button>
    </div>

    <!-- 搜索结果显示区 -->
    <div v-if="searchResults.length > 0" class="search-results-section">
      <div class="section-title">搜索结果</div>
      <div v-for="user in searchResults" :key="user.id" class="contact-item">
        <div class="contact-avatar">
          <img v-if="getAvatarUrl(user.avatar)" :src="getAvatarUrl(user.avatar)" class="avatar-img" />
          <span v-else>{{ user.friend_name?.[0] || '?' }}</span>
        </div>
        <div class="contact-info">
          <div class="contact-name">{{ user.friend_name }}</div>
          <div class="contact-signature">{{ user.signature || '' }}</div>
        </div>
        <button v-if="friendIds.has(user.id)" class="btn-small btn-chat" @click="emit('startChat', user.id)">发消息</button>
        <button v-else-if="sentIds.has(user.id)" class="btn-small btn-sent" disabled>已发送请求</button>
        <button v-else-if="sendingIds.has(user.id)" class="btn-small" disabled>发送中...</button>
        <button v-else class="btn-small" @click="addFriend(user.id)">加好友</button>
      </div>
    </div>

    <!-- 好友申请通知 -->
    <div v-if="friendStore.applyList.length > 0" class="apply-hint" @click="openNotify">
      有 {{ friendStore.applyList.length }} 条好友申请
    </div>

    <!-- 好友列表 -->
    <div class="section-title">我的好友</div>
    <div class="friend-list">
      <div v-if="friendStore.friendList.length === 0" class="empty-tip">暂无好友</div>
      <div
        v-for="f in friendStore.friendList"
        :key="f.id"
        class="contact-item"
        @click="emit('showDetail', f)"
      >
        <div class="contact-avatar">
          <img v-if="getAvatarUrl(f.avatar)" :src="getAvatarUrl(f.avatar)" class="avatar-img" />
          <span v-else>{{ f.remark?.[0] || f.friend_name?.[0] || '?' }}</span>
        </div>
        <div class="contact-info">
          <div class="contact-name">{{ f.remark || f.friend_name }}</div>
          <div class="contact-signature">{{ f.signature || '' }}</div>
        </div>
        <span :class="['online-dot-sm', f.online_statu ? 'online' : '']" :title="f.online_statu ? '在线' : '离线'" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-bar { display: flex; gap: 8px; padding: 10px 16px; }
.search-bar input { flex: 1; padding: 8px 12px; border: 1px solid #ddd; border-radius: 6px; font-size: 13px; outline: none; }
.search-bar input:focus { border-color: #667eea; }
.search-results-section { border-bottom: 1px solid #f0f0f0; padding-bottom: 4px; }
.section-title { font-size: 12px; color: #999; padding: 8px 16px 4px; }
.apply-hint { margin: 4px 16px; padding: 8px 12px; background: #fff3cd; border-radius: 6px; font-size: 13px; color: #856404; cursor: pointer; }
.friend-list { flex: 1; overflow-y: auto; }
.online-dot-sm { width: 8px; height: 8px; border-radius: 50%; background: #ccc; flex-shrink: 0; }
.online-dot-sm.online { background: #2ecc71; }
.btn-chat { background: #667eea; color: #fff; }
.btn-sent { background: #95a5a6; color: #fff; cursor: default; }
</style>
